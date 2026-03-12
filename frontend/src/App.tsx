import { useEffect, useState } from "react";
import "./index.css";
import TopBar from "./components/TopBar";
import AlertInbox from "./components/AlertInbox";
import MapPanel from "./components/MapPanel";
import DetailPanel from "./components/DetailPanel";
import type { Alert, Asset, EventItem } from "./types";

const API_BASE = "http://localhost:8080";

export default function App() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
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
      } catch (error) {
        console.error("Failed to load dashboard data:", error);
      } finally {
        setLoading(false);
      }
    }

    loadData();
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
        <TopBar role="Operator" />
        <main className="loading-state">Loading mission data...</main>
      </div>
    );
  }

  return (
    <div className="app-shell">
      <TopBar role="Operator" />

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