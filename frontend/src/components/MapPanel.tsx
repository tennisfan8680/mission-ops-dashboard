import type { Asset, EventItem } from "../types";

type MapPanelProps = {
  assets: Asset[];
  events: EventItem[];
};

export default function MapPanel({ assets, events }: MapPanelProps) {
  return (
    <section className="panel map-panel">
      <div className="panel-header">
        <h2>Operational Picture</h2>
        <span>Live View</span>
      </div>

      <div className="map-placeholder">
        <div className="grid-overlay" />

        {assets.map((asset, index) => (
          <div
            key={asset.id}
            className="map-marker asset-marker"
            style={{
              top: `${25 + index * 18}%`,
              left: `${20 + index * 22}%`
            }}
            title={`${asset.name} — ${asset.status}`}
          >
            {asset.name}
          </div>
        ))}

        {events.map((event, index) => (
          <div
            key={event.id}
            className="map-marker event-marker"
            style={{
              top: `${18 + index * 20}%`,
              left: `${58 - index * 10}%`
            }}
            title={`${event.type} — ${event.location}`}
          >
            {event.type}
          </div>
        ))}
      </div>
    </section>
  );
}