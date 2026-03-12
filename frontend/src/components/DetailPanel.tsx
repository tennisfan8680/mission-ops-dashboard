import type { Alert } from "../types";

type DetailPanelProps = {
  selectedAlert: Alert | null;
  onAction: (alertId: number, action: "ack" | "escalate" | "close") => void;
};

export default function DetailPanel({ selectedAlert, onAction }: DetailPanelProps) {
  return (
    <section className="panel detail-panel">
      <div className="panel-header">
        <h2>Details</h2>
        <span>Filters + Context</span>
      </div>

      <div className="filters-box">
        <h3>Quick Filters</h3>
        <div className="filter-tags">
          <span>Last 1h</span>
          <span>High Severity</span>
          <span>Maritime</span>
          <span>Assigned to Me</span>
        </div>
      </div>

      <div className="detail-box">
        {selectedAlert ? (
          <>
            <h3>{selectedAlert.title}</h3>
            <p><strong>Severity:</strong> {selectedAlert.severity}</p>
            <p><strong>Status:</strong> {selectedAlert.status}</p>
            <p><strong>Location:</strong> {selectedAlert.location}</p>
            <p><strong>Timestamp:</strong> {selectedAlert.timestamp}</p>
            <p><strong>Description:</strong> {selectedAlert.description}</p>

            <div className="action-row">
              <button onClick={() => onAction(selectedAlert.id, "ack")}>
                Acknowledge
              </button>
              <button onClick={() => onAction(selectedAlert.id, "escalate")}>
                Escalate
              </button>
              <button onClick={() => onAction(selectedAlert.id, "close")}>
                Close
              </button>
            </div>
          </>
        ) : (
          <p>Select an alert to view more details.</p>
        )}
      </div>
    </section>
  );
}