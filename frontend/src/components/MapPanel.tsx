import { CircleMarker, MapContainer, Popup, TileLayer } from "react-leaflet";
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

      <div style={{ height: "580px", borderRadius: "14px", overflow: "hidden" }}>
        <MapContainer
          center={[37.26, -77.38]}
          zoom={9}
          style={{ height: "100%", width: "100%" }}
        >
          <TileLayer
            attribution="© OpenStreetMap contributors"
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          />

          {assets.map((asset) => (
            <CircleMarker
              key={asset.id}
              center={[asset.lat, asset.lon]}
              radius={9}
              pathOptions={{
                color: "#d1d5db",
                fillColor: "#9ca3af",
                fillOpacity: 0.85
              }}
            >
              <Popup>
                <strong>{asset.name}</strong>
                <br />
                {asset.type}
                <br />
                Status: {asset.status}
              </Popup>
            </CircleMarker>
          ))}

          {events.map((event) => (
            <CircleMarker
              key={event.id}
              center={[event.lat, event.lon]}
              radius={7}
              pathOptions={{
                color: "#ef4444",
                fillColor: "#f87171",
                fillOpacity: 0.9
              }}
            >
              <Popup>
                <strong>{event.type}</strong>
                <br />
                {event.location}
                <br />
                {event.details}
              </Popup>
            </CircleMarker>
          ))}
        </MapContainer>
      </div>
    </section>
  );
}