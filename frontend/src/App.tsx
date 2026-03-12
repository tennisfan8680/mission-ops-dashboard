import { useState } from "react";
import "./index.css";
import TopBar from "./components/TopBar";
import AlertInbox from "./components/AlertInbox";
import MapPanel from "./components/MapPanel";
import DetailPanel from "./components/DetailPanel";
import { mockAlerts, mockAssets, mockEvents } from "./data/mockData";
import type { Alert } from "./types";

export default function App() {
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(mockAlerts[0]);

  return (
    <div className="app-shell">
      <TopBar role="Operator" />

      <main className="dashboard-grid">
        <AlertInbox
          alerts={mockAlerts}
          selectedAlertId={selectedAlert?.id ?? null}
          onSelectAlert={setSelectedAlert}
        />

        <MapPanel assets={mockAssets} events={mockEvents} />

        <DetailPanel selectedAlert={selectedAlert} />
      </main>
    </div>
  );
}