package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Location  string `json:"location"`
	Details   string `json:"details"`
}

var alerts = []Alert{
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

var assets = []Asset{
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

var events = []EventItem{
	{
		ID:        201,
		Type:      "Radar Contact",
		Timestamp: "14:10 UTC",
		Location:  "Sector C4",
		Details:   "New contact classified as unknown.",
	},
	{
		ID:        202,
		Type:      "Signal Loss",
		Timestamp: "13:52 UTC",
		Location:  "Relay Tower 7",
		Details:   "Comms quality dropped below threshold.",
	},
	{
		ID:        203,
		Type:      "Route Delay",
		Timestamp: "13:25 UTC",
		Location:  "Route Bravo",
		Details:   "Convoy slowed by route congestion.",
	},
}

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

	alert, found := findAlertByID(id)
	if !found {
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
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "unknown alert action",
		})
		return
	}

	writeJSON(w, http.StatusOK, alert)
}

func main() {
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
		writeJSON(w, http.StatusOK, alerts)
	})

	mux.HandleFunc("/alerts/", handleAlertAction)

	mux.HandleFunc("/assets", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		writeJSON(w, http.StatusOK, assets)
	})

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		writeJSON(w, http.StatusOK, events)
	})

	addr := ":" + port
	log.Printf("backend listening on %s", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
