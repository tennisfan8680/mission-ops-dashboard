export type User = {
  id: string
  email: string
  name: string
  role: 'analyst' | 'operator' | 'commander'
}

export type Asset = {
  id: string
  name: string
  type: string
  status: string
  region: string
  lat: number
  lon: number
  lastSeen: string
}

export type EventItem = {
  id: string
  type: string
  source: string
  region: string
  assetId?: string
  lat: number
  lon: number
  timestamp: string
  payload: {
    confidence?: number
    summary?: string
  }
}

export type AlertItem = {
  id: string
  title: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  status: 'open' | 'acknowledged' | 'closed' | 'escalated'
  region: string
  assetId?: string
  assignedTo?: string
  createdAt: string
  updatedAt: string
  note: string
}
