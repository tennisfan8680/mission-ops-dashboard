export type AlertSeverity = "Low" | "Medium" | "High" | "Critical";
export type AlertStatus = "Open" | "Acknowledged" | "Escalated" | "Closed";

export interface Asset {
  id: number;
  name: string;
  type: string;
  status: string;
  lat: number;
  lon: number;
  lastSeen: string;
}

export interface Alert {
  id: number;
  title: string;
  severity: AlertSeverity;
  status: AlertStatus;
  timestamp: string;
  location: string;
  assetId?: number;
  description: string;
}

export interface EventItem {
  id: number;
  type: string;
  timestamp: string;
  location: string;
  details: string;
}