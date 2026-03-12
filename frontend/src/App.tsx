import { useEffect, useMemo, useState } from 'react'
import type { AlertItem, Asset, EventItem, User } from './types'

const API_BASE = import.meta.env.VITE_API_BASE ?? 'http://localhost:8080'

const demoCreds = [
  { role: 'Analyst', email: 'analyst@missionops.local', password: 'mission123' },
  { role: 'Operator', email: 'operator@missionops.local', password: 'operator123' },
  { role: 'Commander', email: 'commander@missionops.local', password: 'command123' },
]

function App() {
  const [token, setToken] = useState<string>('')
  const [user, setUser] = useState<User | null>(null)
  const [email, setEmail] = useState(demoCreds[1].email)
  const [password, setPassword] = useState(demoCreds[1].password)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const [assets, setAssets] = useState<Asset[]>([])
  const [events, setEvents] = useState<EventItem[]>([])
  const [alerts, setAlerts] = useState<AlertItem[]>([])
  const [selectedRegion, setSelectedRegion] = useState('all')
  const [selectedHours, setSelectedHours] = useState('6')
  const [selectedAlertStatus, setSelectedAlertStatus] = useState('all')
  const [selectedSeverity, setSelectedSeverity] = useState('all')
  const [assignedToMe, setAssignedToMe] = useState(false)

  async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers = new Headers(options.headers)
    headers.set('Content-Type', 'application/json')
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }
    const response = await fetch(`${API_BASE}${path}`, { ...options, headers })
    const body = await response.json().catch(() => ({}))
    if (!response.ok) {
      throw new Error(body.error ?? 'Request failed')
    }
    return body as T
  }

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const res = await fetch(`${API_BASE}/api/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })
      const body = await res.json()
      if (!res.ok) {
        throw new Error(body.error ?? 'Login failed')
      }
      setToken(body.token)
      setUser(body.user)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  async function loadData() {
    if (!token) return
    const regionParam = selectedRegion === 'all' ? '' : `&region=${selectedRegion}`
    const statusParam = selectedAlertStatus === 'all' ? '' : `&status=${selectedAlertStatus}`
    const severityParam = selectedSeverity === 'all' ? '' : `&severity=${selectedSeverity}`
    const assignedParam = assignedToMe ? '&assignedTo=me' : ''

    const [assetData, eventData, alertData] = await Promise.all([
      request<Asset[]>(`/api/assets?status=active${regionParam}`),
      request<EventItem[]>(`/api/events?fromHours=${selectedHours}${regionParam}`),
      request<AlertItem[]>(`/api/alerts?${statusParam.replace('&', '')}${severityParam}${assignedParam}`),
    ])

    setAssets(assetData)
    setEvents(eventData)
    setAlerts(alertData)
  }

  useEffect(() => {
    if (!token) return
    loadData().catch((err) => setError(err instanceof Error ? err.message : 'Load failed'))
    const timer = window.setInterval(() => {
      loadData().catch(() => undefined)
    }, 5000)
    return () => window.clearInterval(timer)
  }, [token, selectedRegion, selectedHours, selectedAlertStatus, selectedSeverity, assignedToMe])

  const regionCounts = useMemo(() => {
    const counts = new Map<string, number>()
    alerts.forEach((alert) => counts.set(alert.region, (counts.get(alert.region) ?? 0) + 1))
    return Array.from(counts.entries())
  }, [alerts])

  const severityCounts = useMemo(() => {
    return ['critical', 'high', 'medium', 'low'].map((level) => ({
      level,
      count: alerts.filter((alert) => alert.severity === level).length,
    }))
  }, [alerts])

  async function actOnAlert(alertId: string, action: 'ack' | 'close' | 'escalate') {
    try {
      await request(`/api/alerts/${alertId}/${action}`, { method: 'POST' })
      await loadData()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action failed')
    }
  }

  if (!token || !user) {
    return (
      <div className="login-shell">
        <div className="login-card">
          <p className="eyebrow">Defense Tech Portfolio Project</p>
          <h1>Mission Ops Dashboard</h1>
          <p className="muted">A mini common operating picture with alert triage, role-based access, and live operational data.</p>

          <form onSubmit={handleLogin} className="login-form">
            <label>
              Email
              <input value={email} onChange={(e) => setEmail(e.target.value)} />
            </label>
            <label>
              Password
              <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
            </label>
            <button disabled={loading}>{loading ? 'Signing in...' : 'Sign in'}</button>
          </form>

          {error ? <div className="error-banner">{error}</div> : null}

          <div className="creds-grid">
            {demoCreds.map((cred) => (
              <button
                key={cred.role}
                className="cred-card"
                onClick={() => {
                  setEmail(cred.email)
                  setPassword(cred.password)
                }}
              >
                <strong>{cred.role}</strong>
                <span>{cred.email}</span>
              </button>
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Operational Command View</p>
          <h1>Mission Ops Dashboard</h1>
        </div>
        <div className="topbar-actions">
          <div className="pill">Signed in as {user.name} ({user.role})</div>
          <button className="ghost" onClick={() => { setToken(''); setUser(null) }}>Log out</button>
        </div>
      </header>

      {error ? <div className="error-banner">{error}</div> : null}

      <section className="summary-grid">
        <article className="card summary-card">
          <h3>Alert Severity Mix</h3>
          {severityCounts.map((item) => (
            <div key={item.level} className="row-between">
              <span className={`severity-badge ${item.level}`}>{item.level}</span>
              <strong>{item.count}</strong>
            </div>
          ))}
        </article>
        <article className="card summary-card">
          <h3>Active Regions</h3>
          {regionCounts.map(([region, count]) => (
            <div key={region} className="row-between">
              <span>{region}</span>
              <strong>{count}</strong>
            </div>
          ))}
        </article>
        <article className="card summary-card">
          <h3>Live Feed</h3>
          <div className="metric">{assets.length} Active Assets</div>
          <div className="metric">{events.length} Recent Events</div>
          <div className="metric">{alerts.length} Alerts</div>
        </article>
      </section>

      <section className="workspace-grid">
        <aside className="card panel">
          <div className="panel-header">
            <h2>Alert Inbox</h2>
            <span>{alerts.length}</span>
          </div>
          <div className="alert-list">
            {alerts.map((alert) => (
              <div className="alert-item" key={alert.id}>
                <div className="row-between">
                  <strong>{alert.title}</strong>
                  <span className={`severity-badge ${alert.severity}`}>{alert.severity}</span>
                </div>
                <p className="muted compact">{alert.region} · {alert.status} · {new Date(alert.updatedAt).toLocaleTimeString()}</p>
                <p>{alert.note}</p>
                <div className="action-row">
                  <button onClick={() => actOnAlert(alert.id, 'ack')} disabled={user.role === 'analyst'}>Ack</button>
                  <button onClick={() => actOnAlert(alert.id, 'close')} disabled={user.role === 'analyst'}>Close</button>
                  <button onClick={() => actOnAlert(alert.id, 'escalate')} disabled={user.role !== 'commander'}>Escalate</button>
                </div>
              </div>
            ))}
          </div>
        </aside>

        <main className="card panel map-panel">
          <div className="panel-header">
            <h2>Common Operating Picture</h2>
            <span>Approximate operational map</span>
          </div>
          <div className="map-box">
            {assets.map((asset) => (
              <div
                key={asset.id}
                className="map-point asset"
                style={{ left: `${40 + (asset.lon + 78) * 70}%`, top: `${80 - (asset.lat - 38.7) * 130}%` }}
                title={`${asset.name} (${asset.type})`}
              >
                <span>{asset.name}</span>
              </div>
            ))}
            {events.slice(0, 12).map((event) => (
              <div
                key={event.id}
                className="map-point event"
                style={{ left: `${40 + (event.lon + 78) * 70}%`, top: `${80 - (event.lat - 38.7) * 130}%` }}
                title={event.payload?.summary ?? event.type}
              />
            ))}
          </div>
          <div className="legend-row">
            <span><i className="dot asset" /> Asset</span>
            <span><i className="dot event" /> Event</span>
          </div>
        </main>

        <aside className="card panel">
          <div className="panel-header">
            <h2>Filters & Recent Events</h2>
          </div>

          <div className="filters">
            <label>
              Region
              <select value={selectedRegion} onChange={(e) => setSelectedRegion(e.target.value)}>
                <option value="all">All</option>
                <option value="north">North</option>
                <option value="east">East</option>
                <option value="south">South</option>
              </select>
            </label>
            <label>
              Time Window
              <select value={selectedHours} onChange={(e) => setSelectedHours(e.target.value)}>
                <option value="1">Last 1 hour</option>
                <option value="6">Last 6 hours</option>
                <option value="24">Last 24 hours</option>
              </select>
            </label>
            <label>
              Alert Status
              <select value={selectedAlertStatus} onChange={(e) => setSelectedAlertStatus(e.target.value)}>
                <option value="all">All</option>
                <option value="open">Open</option>
                <option value="acknowledged">Acknowledged</option>
                <option value="escalated">Escalated</option>
                <option value="closed">Closed</option>
              </select>
            </label>
            <label>
              Severity
              <select value={selectedSeverity} onChange={(e) => setSelectedSeverity(e.target.value)}>
                <option value="all">All</option>
                <option value="critical">Critical</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </label>
            <label className="checkbox-row">
              <input type="checkbox" checked={assignedToMe} onChange={(e) => setAssignedToMe(e.target.checked)} />
              Assigned to me
            </label>
          </div>

          <div className="event-feed">
            {events.slice(0, 8).map((event) => (
              <div key={event.id} className="feed-item">
                <strong>{event.type.replaceAll('_', ' ')}</strong>
                <p>{event.payload?.summary ?? 'Operational update received.'}</p>
                <span>{new Date(event.timestamp).toLocaleTimeString()} · {event.source}</span>
              </div>
            ))}
          </div>
        </aside>
      </section>
    </div>
  )
}

export default App
