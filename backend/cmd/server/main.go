package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "os"
    "sort"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    _ "github.com/jackc/pgx/v5/stdlib"
)

type application struct {
    db        *sql.DB
    jwtSecret []byte
}

type userClaims struct {
    UserID string `json:"userId"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

type user struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
    Role  string `json:"role"`
}

type asset struct {
    ID       string    `json:"id"`
    Name     string    `json:"name"`
    Type     string    `json:"type"`
    Status   string    `json:"status"`
    Region   string    `json:"region"`
    Lat      float64   `json:"lat"`
    Lon      float64   `json:"lon"`
    LastSeen time.Time `json:"lastSeen"`
}

type event struct {
    ID        string          `json:"id"`
    Type      string          `json:"type"`
    Source    string          `json:"source"`
    Region    string          `json:"region"`
    AssetID   *string         `json:"assetId"`
    Lat       float64         `json:"lat"`
    Lon       float64         `json:"lon"`
    Timestamp time.Time       `json:"timestamp"`
    Payload   json.RawMessage `json:"payload"`
}

type alert struct {
    ID         string     `json:"id"`
    Title      string     `json:"title"`
    Severity   string     `json:"severity"`
    Status     string     `json:"status"`
    Region     string     `json:"region"`
    AssetID    *string    `json:"assetId"`
    AssignedTo *string    `json:"assignedTo"`
    CreatedAt  time.Time  `json:"createdAt"`
    UpdatedAt  time.Time  `json:"updatedAt"`
    Note       string     `json:"note"`
}

type loginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type healthResponse struct {
    Status    string `json:"status"`
    Timestamp string `json:"timestamp"`
}

func main() {
    port := envOrDefault("PORT", "8080")
    databaseURL := envOrDefault("DATABASE_URL", "postgres://missionops:missionops@localhost:5432/missionops?sslmode=disable")
    jwtSecret := envOrDefault("JWT_SECRET", "dev-secret-change-me")

    db, err := sql.Open("pgx", databaseURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        log.Fatal(err)
    }

    app := &application{db: db, jwtSecret: []byte(jwtSecret)}

    go app.startSimulator()

    mux := http.NewServeMux()
    mux.HandleFunc("/health", app.handleHealth)
    mux.HandleFunc("/api/login", app.handleLogin)
    mux.HandleFunc("/api/me", app.withAuth(app.handleMe))
    mux.HandleFunc("/api/assets", app.withAuth(app.handleAssets))
    mux.HandleFunc("/api/events", app.withAuth(app.handleEvents))
    mux.HandleFunc("/api/alerts", app.withAuth(app.handleAlerts))
    mux.HandleFunc("/api/alerts/", app.withAuth(app.handleAlertAction))

    handler := app.withCORS(app.withLogging(mux))

    server := &http.Server{
        Addr:              ":" + port,
        Handler:           handler,
        ReadHeaderTimeout: 5 * time.Second,
    }

    log.Printf("backend listening on %s", server.Addr)
    log.Fatal(server.ListenAndServe())
}

func envOrDefault(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}

func (app *application) handleHealth(w http.ResponseWriter, r *http.Request) {
    app.writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Timestamp: time.Now().UTC().Format(time.RFC3339)})
}

func (app *application) handleLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        app.methodNotAllowed(w)
        return
    }

    var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        app.badRequest(w, "invalid JSON payload")
        return
    }

    var u user
    var passwordHash string
    err := app.db.QueryRow(`
        SELECT id, email, name, role, password_hash
        FROM users
        WHERE email = $1
    `, strings.ToLower(strings.TrimSpace(req.Email))).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &passwordHash)
    if err != nil {
        app.unauthorized(w, "invalid email or password")
        return
    }

    if passwordHash != req.Password {
        app.unauthorized(w, "invalid email or password")
        return
    }

    token, err := app.makeToken(u)
    if err != nil {
        app.serverError(w, err)
        return
    }

    app.writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}

func (app *application) handleMe(w http.ResponseWriter, r *http.Request, currentUser user) {
    app.writeJSON(w, http.StatusOK, currentUser)
}

func (app *application) handleAssets(w http.ResponseWriter, r *http.Request, _ user) {
    q := r.URL.Query()
    region := strings.TrimSpace(q.Get("region"))
    types := splitCSV(q.Get("types"))
    status := strings.TrimSpace(q.Get("status"))

    rows, err := app.db.Query(`
        SELECT id, name, type, status, region, lat, lon, last_seen
        FROM assets
    `)
    if err != nil {
        app.serverError(w, err)
        return
    }
    defer rows.Close()

    var assets []asset
    for rows.Next() {
        var a asset
        if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Status, &a.Region, &a.Lat, &a.Lon, &a.LastSeen); err != nil {
            app.serverError(w, err)
            return
        }
        if region != "" && !strings.EqualFold(a.Region, region) {
            continue
        }
        if status != "" && !strings.EqualFold(a.Status, status) {
            continue
        }
        if len(types) > 0 && !containsFold(types, a.Type) {
            continue
        }
        assets = append(assets, a)
    }

    sort.Slice(assets, func(i, j int) bool { return assets[i].Name < assets[j].Name })
    app.writeJSON(w, http.StatusOK, assets)
}

func (app *application) handleEvents(w http.ResponseWriter, r *http.Request, _ user) {
    q := r.URL.Query()
    region := strings.TrimSpace(q.Get("region"))
    eventTypes := splitCSV(q.Get("types"))
    from, to := parseTimeRange(q.Get("fromHours"))

    rows, err := app.db.Query(`
        SELECT id, type, source, region, asset_id, lat, lon, ts, payload_json
        FROM events
        WHERE ts BETWEEN $1 AND $2
        ORDER BY ts DESC
        LIMIT 250
    `, from, to)
    if err != nil {
        app.serverError(w, err)
        return
    }
    defer rows.Close()

    var events []event
    for rows.Next() {
        var e event
        if err := rows.Scan(&e.ID, &e.Type, &e.Source, &e.Region, &e.AssetID, &e.Lat, &e.Lon, &e.Timestamp, &e.Payload); err != nil {
            app.serverError(w, err)
            return
        }
        if region != "" && !strings.EqualFold(e.Region, region) {
            continue
        }
        if len(eventTypes) > 0 && !containsFold(eventTypes, e.Type) {
            continue
        }
        events = append(events, e)
    }

    app.writeJSON(w, http.StatusOK, events)
}

func (app *application) handleAlerts(w http.ResponseWriter, r *http.Request, currentUser user) {
    if r.Method != http.MethodGet {
        app.methodNotAllowed(w)
        return
    }

    q := r.URL.Query()
    statusFilter := strings.TrimSpace(q.Get("status"))
    severityFilter := strings.TrimSpace(q.Get("severity"))
    assignedToMe := q.Get("assignedTo") == "me"

    rows, err := app.db.Query(`
        SELECT id, title, severity, status, region, asset_id, assigned_to, created_at, updated_at, note
        FROM alerts
        ORDER BY updated_at DESC
        LIMIT 250
    `)
    if err != nil {
        app.serverError(w, err)
        return
    }
    defer rows.Close()

    var alerts []alert
    for rows.Next() {
        var a alert
        if err := rows.Scan(&a.ID, &a.Title, &a.Severity, &a.Status, &a.Region, &a.AssetID, &a.AssignedTo, &a.CreatedAt, &a.UpdatedAt, &a.Note); err != nil {
            app.serverError(w, err)
            return
        }
        if statusFilter != "" && !strings.EqualFold(a.Status, statusFilter) {
            continue
        }
        if severityFilter != "" && !strings.EqualFold(a.Severity, severityFilter) {
            continue
        }
        if assignedToMe {
            if a.AssignedTo == nil || *a.AssignedTo != currentUser.ID {
                continue
            }
        }
        alerts = append(alerts, a)
    }

    app.writeJSON(w, http.StatusOK, alerts)
}

func (app *application) handleAlertAction(w http.ResponseWriter, r *http.Request, currentUser user) {
    if r.Method != http.MethodPost {
        app.methodNotAllowed(w)
        return
    }

    if currentUser.Role == "analyst" {
        app.forbidden(w, "analysts have view-only access")
        return
    }

    path := strings.TrimPrefix(r.URL.Path, "/api/alerts/")
    parts := strings.Split(path, "/")
    if len(parts) != 2 {
        app.notFound(w)
        return
    }
    alertID := parts[0]
    action := parts[1]

    switch action {
    case "ack":
        if err := app.updateAlert(alertID, "acknowledged", &currentUser.ID, currentUser.ID, "acknowledged"); err != nil {
            app.handleActionError(w, err)
            return
        }
    case "close":
        if err := app.updateAlert(alertID, "closed", &currentUser.ID, currentUser.ID, "closed"); err != nil {
            app.handleActionError(w, err)
            return
        }
    case "escalate":
        if currentUser.Role != "commander" {
            app.forbidden(w, "only commanders can escalate alerts")
            return
        }
        if err := app.updateAlert(alertID, "escalated", &currentUser.ID, currentUser.ID, "escalated"); err != nil {
            app.handleActionError(w, err)
            return
        }
    default:
        app.notFound(w)
        return
    }

    app.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (app *application) updateAlert(alertID, newStatus string, assignedTo *string, actorID, action string) error {
    tx, err := app.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    result, err := tx.Exec(`
        UPDATE alerts
        SET status = $1, assigned_to = $2, updated_at = NOW()
        WHERE id = $3
    `, newStatus, assignedTo, alertID)
    if err != nil {
        return err
    }
    if result.RowsAffected() == 0 {
        return sql.ErrNoRows
    }

    _, err = tx.Exec(`
        INSERT INTO audit_log (id, actor_id, action, target_type, target_id)
        VALUES ($1, $2, $3, 'alert', $4)
    `, uuid.NewString(), actorID, action, alertID)
    if err != nil {
        return err
    }

    return tx.Commit()
}

func (app *application) handleActionError(w http.ResponseWriter, err error) {
    if errors.Is(err, sql.ErrNoRows) {
        app.notFound(w)
        return
    }
    app.serverError(w, err)
}

func (app *application) startSimulator() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    types := []string{"sensor_anomaly", "geofence_breach", "comms_loss", "route_deviation"}
    severities := []string{"low", "medium", "high"}
    sources := []string{"drone-feed", "ops-console", "perimeter-sensor", "satellite-ingest"}

    for range ticker.C {
        assets, err := app.listAssets()
        if err != nil || len(assets) == 0 {
            continue
        }
        a := assets[rand.Intn(len(assets))]
        eventType := types[rand.Intn(len(types))]
        source := sources[rand.Intn(len(sources))]
        latJitter := (rand.Float64() - 0.5) * 0.08
        lonJitter := (rand.Float64() - 0.5) * 0.08
        eventID := uuid.NewString()

        payload := map[string]any{
            "confidence": 80 + rand.Intn(19),
            "summary":    fmt.Sprintf("%s detected near %s", strings.ReplaceAll(eventType, "_", " "), a.Name),
        }
        payloadJSON, _ := json.Marshal(payload)

        _, err = app.db.Exec(`
            INSERT INTO events (id, type, source, region, asset_id, lat, lon, ts, payload_json)
            VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), $8)
        `, eventID, eventType, source, a.Region, a.ID, a.Lat+latJitter, a.Lon+lonJitter, payloadJSON)
        if err != nil {
            continue
        }

        // Roughly every other event creates an alert.
        if rand.Intn(2) == 0 {
            severity := severities[rand.Intn(len(severities))]
            title := strings.Title(strings.ReplaceAll(eventType, "_", " "))
            note := fmt.Sprintf("Auto-generated from %s for %s.", source, a.Name)
            _, _ = app.db.Exec(`
                INSERT INTO alerts (id, title, severity, status, region, asset_id, assigned_to, created_at, updated_at, note)
                VALUES ($1, $2, $3, 'open', $4, $5, NULL, NOW(), NOW(), $6)
            `, uuid.NewString(), title, severity, a.Region, a.ID, note)
        }
    }
}

func (app *application) listAssets() ([]asset, error) {
    rows, err := app.db.Query(`SELECT id, name, type, status, region, lat, lon, last_seen FROM assets`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var assets []asset
    for rows.Next() {
        var a asset
        if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Status, &a.Region, &a.Lat, &a.Lon, &a.LastSeen); err != nil {
            return nil, err
        }
        assets = append(assets, a)
    }
    return assets, nil
}

func (app *application) makeToken(u user) (string, error) {
    claims := userClaims{
        UserID: u.ID,
        Email:  u.Email,
        Role:   u.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   u.ID,
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(app.jwtSecret)
}

func (app *application) withAuth(next func(http.ResponseWriter, *http.Request, user)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            app.unauthorized(w, "missing bearer token")
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims := &userClaims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
            return app.jwtSecret, nil
        })
        if err != nil || !token.Valid {
            app.unauthorized(w, "invalid bearer token")
            return
        }

        currentUser := user{ID: claims.UserID, Email: claims.Email, Role: claims.Role}
        err = app.db.QueryRow(`SELECT name FROM users WHERE id = $1`, claims.UserID).Scan(&currentUser.Name)
        if err != nil {
            app.unauthorized(w, "unknown user")
            return
        }

        next(w, r, currentUser)
    }
}

func (app *application) withLogging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        started := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started))
    })
}

func (app *application) withCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func splitCSV(value string) []string {
    if value == "" {
        return nil
    }
    raw := strings.Split(value, ",")
    out := make([]string, 0, len(raw))
    for _, item := range raw {
        item = strings.TrimSpace(item)
        if item != "" {
            out = append(out, item)
        }
    }
    return out
}

func containsFold(items []string, value string) bool {
    for _, item := range items {
        if strings.EqualFold(item, value) {
            return true
        }
    }
    return false
}

func parseTimeRange(fromHours string) (time.Time, time.Time) {
    to := time.Now().UTC()
    switch fromHours {
    case "1":
        return to.Add(-1 * time.Hour), to
    case "6":
        return to.Add(-6 * time.Hour), to
    case "24":
        return to.Add(-24 * time.Hour), to
    default:
        return to.Add(-6 * time.Hour), to
    }
}

func (app *application) writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func (app *application) badRequest(w http.ResponseWriter, message string) {
    app.writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
}

func (app *application) unauthorized(w http.ResponseWriter, message string) {
    app.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": message})
}

func (app *application) forbidden(w http.ResponseWriter, message string) {
    app.writeJSON(w, http.StatusForbidden, map[string]string{"error": message})
}

func (app *application) notFound(w http.ResponseWriter) {
    app.writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func (app *application) methodNotAllowed(w http.ResponseWriter) {
    app.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (app *application) serverError(w http.ResponseWriter, err error) {
    log.Println("server error:", err)
    app.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
