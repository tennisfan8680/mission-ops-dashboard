import type { Alert } from "../types";

type AlertInboxProps = {
  alerts: Alert[];
  selectedAlertId: number | null;
  onSelectAlert: (alert: Alert) => void;
};

export default function AlertInbox({
  alerts,
  selectedAlertId,
  onSelectAlert
}: AlertInboxProps) {
  return (
    <section className="panel alert-panel">
      <div className="panel-header">
        <h2>Alert Inbox</h2>
        <span>{alerts.length} alerts</span>
      </div>

      <div className="alert-list">
        {alerts.map((alert) => (
          <button
            key={alert.id}
            className={`alert-card ${selectedAlertId === alert.id ? "selected" : ""}`}
            onClick={() => onSelectAlert(alert)}
          >
            <div className="alert-card-top">
              <span className={`severity ${alert.severity.toLowerCase()}`}>
                {alert.severity}
              </span>
              <span className="alert-status">{alert.status}</span>
            </div>

            <h3>{alert.title}</h3>
            <p>{alert.location}</p>
            <small>{alert.timestamp}</small>
          </button>
        ))}
      </div>
    </section>
  );
}