package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type Alert struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	Location    string `json:"location"`
	AssetID     int    `json:"assetId,omitempty"`
	Description string `json:"description"`
}

type Asset struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Status   string  `json:"status"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	LastSeen string  `json:"lastSeen"`
}

type EventItem struct {
	ID        int     `json:"id"`
	Type      string  `json:"type"`
	Timestamp string  `json:"timestamp"`
	Location  string  `json:"location"`
	Details   string  `json:"details"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
}

type DashboardSnapshot struct {
	Alerts      []Alert     `json:"alerts"`
	Assets      []Asset     `json:"assets"`
	Events      []EventItem `json:"events"`
	LastUpdated string      `json:"lastUpdated"`
}

var (
	dataMu sync.Mutex

	alerts = []Alert{
		{
			ID:          1,
			Title:       "Unidentified vessel near sector C4",
			Severity:    "Critical",
			Status:      "Open",
			Timestamp:   "2026-03-12T14:10:00Z",
			Location:    "Sector C4",
			AssetID:     101,
			Description: "Radar picked up an unidentified vessel moving toward a restricted maritime corridor.",
		},
		{
			ID:          2,
			Title:       "Comms degradation at relay tower",
			Severity:    "High",
			Status:      "Acknowledged",
			Timestamp:   "2026-03-12T13:52:00Z",
			Location:    "Relay Tower 7",
			Description: "Packet loss exceeded threshold for 8 minutes.",
		},
		{
			ID:          3,
			Title:       "Supply convoy delayed",
			Severity:    "Medium",
			Status:      "Open",
			Timestamp:   "2026-03-12T13:25:00Z",
			Location:    "Route Bravo",
			Description: "Convoy ETA slipped by 22 minutes due to weather and route congestion.",
		},
		{
			ID:          4,
			Title:       "Thermal anomaly detected",
			Severity:    "High",
			Status:      "Escalated",
			Timestamp:   "2026-03-12T12:58:00Z",
			Location:    "Grid A2",
			Description: "Persistent thermal signature detected near a monitored boundary.",
		},
	}

	assets = []Asset{
		{
			ID:       101,
			Name:     "Patrol-Alpha",
			Type:     "Maritime Patrol",
			Status:   "Active",
			Lat:      36.91,
			Lon:      -75.98,
			LastSeen: "2 min ago",
		},
		{
			ID:       102,
			Name:     "Drone-07",
			Type:     "Recon UAV",
			Status:   "Monitoring",
			Lat:      36.88,
			Lon:      -76.05,
			LastSeen: "30 sec ago",
		},
		{
			ID:       103,
			Name:     "Convoy-Bravo",
			Type:     "Ground Logistics",
			Status:   "Delayed",
			Lat:      36.95,
			Lon:      -76.02,
			LastSeen: "4 min ago",
		},
	}

	events = []EventItem{
		{
			ID:        201,
			Type:      "Radar Contact",
			Timestamp: "14:10 UTC",
			Location:  "Training Sector North",
			Details:   "Unscheduled radar contact detected near inland monitoring sector.",
			Lat:       37.270,
			Lon:       -77.420,
		},
		{
			ID:        202,
			Type:      "Signal Loss",
			Timestamp: "13:52 UTC",
			Location:  "Relay Tower East",
			Details:   "Communications quality dropped below threshold near relay infrastructure.",
			Lat:       37.355,
			Lon:       -77.180,
		},
		{
			ID:        203,
			Type:      "Route Delay",
			Timestamp: "13:25 UTC",
			Location:  "Logistics Corridor West",
			Details:   "Ground convoy slowed due to route congestion.",
			Lat:       37.180,
			Lon:       -77.590,
		},
	}

	nextAlertID = 5
	nextEventID = 204

	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func findAlertByID(id int) (*Alert, bool) {
	for i := range alerts {
		if alerts[i].ID == id {
			return &alerts[i], true
		}
	}
	return nil, false
}

func copyAlerts() []Alert {
	out := make([]Alert, len(alerts))
	copy(out, alerts)
	return out
}

func copyAssets() []Asset {
	out := make([]Asset, len(assets))
	copy(out, assets)
	return out
}

func copyEvents() []EventItem {
	out := make([]EventItem, len(events))
	copy(out, events)
	return out
}

func getSnapshotLocked() DashboardSnapshot {
	return DashboardSnapshot{
		Alerts:      copyAlerts(),
		Assets:      copyAssets(),
		Events:      copyEvents(),
		LastUpdated: time.Now().Format("3:04:05 PM"),
	}
}

func broadcastSnapshot() {
	dataMu.Lock()
	snapshot := getSnapshotLocked()
	dataMu.Unlock()

	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteJSON(snapshot)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	dataMu.Lock()
	initialSnapshot := getSnapshotLocked()
	dataMu.Unlock()

	if err := conn.WriteJSON(initialSnapshot); err != nil {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			conn.Close()
			break
		}
	}
}

func handleAlertAction(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/alerts/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid alert action path",
		})
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid alert id",
		})
		return
	}

	action := parts[1]

	dataMu.Lock()
	alert, found := findAlertByID(id)
	if !found {
		dataMu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "alert not found",
		})
		return
	}

	switch action {
	case "ack":
		alert.Status = "Acknowledged"
	case "escalate":
		alert.Status = "Escalated"
	case "close":
		alert.Status = "Closed"
	default:
		dataMu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "unknown alert action",
		})
		return
	}

	updatedAlert := *alert
	dataMu.Unlock()

	writeJSON(w, http.StatusOK, updatedAlert)
	go broadcastSnapshot()
}

func addSimulatedEvent() {
	eventTemplates := []struct {
		EventType string
		Location  string
		Details   string
		Lat       float64
		Lon       float64
	}{
		{
			EventType: "Radar Contact",
			Location:  "Training Sector North",
			Details:   "Unscheduled radar contact detected near monitored inland zone.",
			Lat:       37.320,
			Lon:       -77.350,
		},
		{
			EventType: "Signal Loss",
			Location:  "Relay Tower East",
			Details:   "Temporary signal degradation detected near communications tower.",
			Lat:       37.360,
			Lon:       -77.160,
		},
		{
			EventType: "Route Delay",
			Location:  "Logistics Corridor West",
			Details:   "Ground movement slowed along logistics corridor.",
			Lat:       37.170,
			Lon:       -77.600,
		},
		{
			EventType: "Thermal Spike",
			Location:  "Grid B7",
			Details:   "Short-duration thermal activity detected near monitoring point.",
			Lat:       37.245,
			Lon:       -77.470,
		},
		{
			EventType: "AIS Anomaly",
			Location:  "Command Zone South",
			Details:   "Asset tracking behavior inconsistent with expected route.",
			Lat:       37.110,
			Lon:       -77.280,
		},
		{
			EventType: "Perimeter Trigger",
			Location:  "Outer Security Ring",
			Details:   "Boundary monitoring system registered an unexpected trigger.",
			Lat:       37.410,
			Lon:       -77.510,
		},
	}

	template := eventTemplates[rand.Intn(len(eventTemplates))]

	latJitter := (rand.Float64() - 0.5) * 0.03
	lonJitter := (rand.Float64() - 0.5) * 0.03

	newEvent := EventItem{
		ID:        nextEventID,
		Type:      template.EventType,
		Timestamp: time.Now().UTC().Format("15:04:05 MST"),
		Location:  template.Location,
		Details:   template.Details,
		Lat:       template.Lat + latJitter,
		Lon:       template.Lon + lonJitter,
	}

	nextEventID++
	events = append([]EventItem{newEvent}, events...)

	if len(events) > 12 {
		events = events[:12]
	}
}

func addSimulatedAlert() {
	alertTemplates := []struct {
		Title       string
		Severity    string
		Location    string
		Description string
	}{
		{
			Title:       "New unidentified surface contact",
			Severity:    "High",
			Location:    "Sector E2",
			Description: "Surface contact deviated from expected maritime traffic pattern.",
		},
		{
			Title:       "Recon drone telemetry drift",
			Severity:    "Medium",
			Location:    "Flight Corridor North",
			Description: "Drone telemetry indicates minor deviation from assigned route.",
		},
		{
			Title:       "Communications interruption detected",
			Severity:    "High",
			Location:    "Relay Network East",
			Description: "Brief communications interruption observed across monitored relay infrastructure.",
		},
		{
			Title:       "Potential perimeter breach indicator",
			Severity:    "Critical",
			Location:    "Grid F4",
			Description: "Multiple sensor triggers detected near protected perimeter boundary.",
		},
	}

	template := alertTemplates[rand.Intn(len(alertTemplates))]

	newAlert := Alert{
		ID:          nextAlertID,
		Title:       template.Title,
		Severity:    template.Severity,
		Status:      "Open",
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Location:    template.Location,
		Description: template.Description,
	}

	nextAlertID++
	alerts = append([]Alert{newAlert}, alerts...)
	if len(alerts) > 12 {
		alerts = alerts[:12]
	}
}

func moveAssetsSlightly() {
	for i := range assets {
		latShift := (rand.Float64() - 0.5) * 0.01
		lonShift := (rand.Float64() - 0.5) * 0.01

		assets[i].Lat += latShift
		assets[i].Lon += lonShift
		assets[i].LastSeen = "just now"
	}
}

func startSimulator() {
	go func() {
		ticker := time.NewTicker(6 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			dataMu.Lock()
			moveAssetsSlightly()

			if rand.Intn(10) < 7 {
				addSimulatedEvent()
			}

			if rand.Intn(100) < 35 {
				addSimulatedAlert()
			}
			dataMu.Unlock()

			broadcastSnapshot()
		}
	}()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	startSimulator()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		resp := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		dataMu.Lock()
		alertsCopy := copyAlerts()
		dataMu.Unlock()

		writeJSON(w, http.StatusOK, alertsCopy)
	})

	mux.HandleFunc("/alerts/", handleAlertAction)

	mux.HandleFunc("/assets", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		dataMu.Lock()
		assetsCopy := copyAssets()
		dataMu.Unlock()

		writeJSON(w, http.StatusOK, assetsCopy)
	})

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		dataMu.Lock()
		eventsCopy := copyEvents()
		dataMu.Unlock()

		writeJSON(w, http.StatusOK, eventsCopy)
	})

	mux.HandleFunc("/ws", handleWebSocket)

	addr := ":" + port
	log.Printf("backend listening on %s", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
