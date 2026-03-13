import { useEffect, useRef, useState } from "react";
import "./index.css";
import TopBar from "./components/TopBar";
import AlertInbox from "./components/AlertInbox";
import MapPanel from "./components/MapPanel";
import DetailPanel from "./components/DetailPanel";
import type { Alert, Asset, EventItem } from "./types";

const API_BASE = "http://localhost:8080";
const WS_URL = "ws://localhost:8080/ws";

type DashboardSnapshot = {
  alerts: Alert[];
  assets: Asset[];
  events: EventItem[];
  lastUpdated: string;
};

export default function App() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<string>("Connecting...");
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    async function loadInitialData() {
      try {
        const [alertsRes, assetsRes, eventsRes] = await Promise.all([
          fetch(`${API_BASE}/alerts`),
          fetch(`${API_BASE}/assets`),
          fetch(`${API_BASE}/events`)
        ]);

        const alertsData: Alert[] = await alertsRes.json();
        const assetsData: Asset[] = await assetsRes.json();
        const eventsData: EventItem[] = await eventsRes.json();

        setAlerts(alertsData);
        setAssets(assetsData);
        setEvents(eventsData);
        setSelectedAlert(alertsData[0] ?? null);
        setLastUpdated(new Date().toLocaleTimeString());
      } catch (error) {
        console.error("Initial load failed:", error);
      } finally {
        setLoading(false);
      }
    }

    loadInitialData();

    const socket = new WebSocket(WS_URL);
    socketRef.current = socket;

    socket.onopen = () => {
      console.log("WebSocket connected");
      setLastUpdated("Live");
    };

    socket.onmessage = (event) => {
      try {
        const snapshot: DashboardSnapshot = JSON.parse(event.data);

        setAlerts(snapshot.alerts);
        setAssets(snapshot.assets);
        setEvents(snapshot.events);
        setLastUpdated(snapshot.lastUpdated);

        setSelectedAlert((currentSelected) => {
          if (!currentSelected) {
            return snapshot.alerts[0] ?? null;
          }

          const refreshedSelected = snapshot.alerts.find(
            (alert) => alert.id === currentSelected.id
          );

          return refreshedSelected ?? snapshot.alerts[0] ?? null;
        });
      } catch (error) {
        console.error("Failed to parse WebSocket message:", error);
      }
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
      setLastUpdated("Connection issue");
    };

    socket.onclose = () => {
      console.log("WebSocket disconnected");
      setLastUpdated("Disconnected");
    };

    return () => {
      socket.close();
    };
  }, []);

  async function handleAlertAction(
    alertId: number,
    action: "ack" | "escalate" | "close"
  ) {
    try {
      const response = await fetch(`${API_BASE}/alerts/${alertId}/${action}`, {
        method: "POST"
      });

      if (!response.ok) {
        throw new Error(`Failed to ${action} alert`);
      }

      const updatedAlert: Alert = await response.json();

      setAlerts((prevAlerts) =>
        prevAlerts.map((alert) =>
          alert.id === updatedAlert.id ? updatedAlert : alert
        )
      );

      setSelectedAlert(updatedAlert);
    } catch (error) {
      console.error("Alert action failed:", error);
    }
  }

  if (loading) {
    return (
      <div className="app-shell">
        <TopBar role="Operator" lastUpdated="Loading..." />
        <main className="loading-state">Loading mission data...</main>
      </div>
    );
  }

  return (
    <div className="app-shell">
      <TopBar role="Operator" lastUpdated={lastUpdated} />

      <main className="dashboard-grid">
        <AlertInbox
          alerts={alerts}
          selectedAlertId={selectedAlert?.id ?? null}
          onSelectAlert={setSelectedAlert}
        />

        <MapPanel assets={assets} events={events} />

        <DetailPanel
          selectedAlert={selectedAlert}
          onAction={handleAlertAction}
        />
      </main>
    </div>
  );
}