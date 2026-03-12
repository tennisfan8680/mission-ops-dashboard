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
          center={[36.91, -75.98]}
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
              radius={10}
              pathOptions={{ color: "#3b82f6", fillColor: "#60a5fa", fillOpacity: 0.8 }}
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

          {events.map((event, index) => (
            <CircleMarker
              key={event.id}
              center={[36.9 + index * 0.02, -76.0 + index * 0.02]}
              radius={8}
              pathOptions={{ color: "#ef4444", fillColor: "#f87171", fillOpacity: 0.8 }}
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