import type { Alert, Asset, EventItem } from "../types";

export const mockAlerts: Alert[] = [
  {
    id: 1,
    title: "Unidentified vessel near sector C4",
    severity: "Critical",
    status: "Open",
    timestamp: "2026-03-12T14:10:00Z",
    location: "Sector C4",
    assetId: 101,
    description: "Radar picked up an unidentified vessel moving toward a restricted maritime corridor."
  },
  {
    id: 2,
    title: "Comms degradation at relay tower",
    severity: "High",
    status: "Acknowledged",
    timestamp: "2026-03-12T13:52:00Z",
    location: "Relay Tower 7",
    description: "Packet loss exceeded threshold for 8 minutes."
  },
  {
    id: 3,
    title: "Supply convoy delayed",
    severity: "Medium",
    status: "Open",
    timestamp: "2026-03-12T13:25:00Z",
    location: "Route Bravo",
    description: "Convoy ETA slipped by 22 minutes due to weather and route congestion."
  },
  {
    id: 4,
    title: "Thermal anomaly detected",
    severity: "High",
    status: "Escalated",
    timestamp: "2026-03-12T12:58:00Z",
    location: "Grid A2",
    description: "Persistent thermal signature detected near a monitored boundary."
  }
];

export const mockAssets: Asset[] = [
  {
    id: 101,
    name: "Patrol-Alpha",
    type: "Maritime Patrol",
    status: "Active",
    lat: 36.91,
    lon: -75.98,
    lastSeen: "2 min ago"
  },
  {
    id: 102,
    name: "Drone-07",
    type: "Recon UAV",
    status: "Monitoring",
    lat: 36.88,
    lon: -76.05,
    lastSeen: "30 sec ago"
  },
  {
    id: 103,
    name: "Convoy-Bravo",
    type: "Ground Logistics",
    status: "Delayed",
    lat: 36.95,
    lon: -76.02,
    lastSeen: "4 min ago"
  }
];

export const mockEvents: EventItem[] = [
  {
    id: 201,
    type: "Radar Contact",
    timestamp: "14:10 UTC",
    location: "Sector C4",
    details: "New contact classified as unknown."
  },
  {
    id: 202,
    type: "Signal Loss",
    timestamp: "13:52 UTC",
    location: "Relay Tower 7",
    details: "Comms quality dropped below threshold."
  },
  {
    id: 203,
    type: "Route Delay",
    timestamp: "13:25 UTC",
    location: "Route Bravo",
    details: "Convoy slowed by route congestion."
  }
];