CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  role TEXT NOT NULL,
  password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS assets (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  status TEXT NOT NULL,
  region TEXT NOT NULL,
  lat DOUBLE PRECISION NOT NULL,
  lon DOUBLE PRECISION NOT NULL,
  last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS events (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  source TEXT NOT NULL,
  region TEXT NOT NULL,
  asset_id TEXT REFERENCES assets(id) ON DELETE SET NULL,
  lat DOUBLE PRECISION NOT NULL,
  lon DOUBLE PRECISION NOT NULL,
  ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  payload_json JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS alerts (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  severity TEXT NOT NULL,
  status TEXT NOT NULL,
  region TEXT NOT NULL,
  asset_id TEXT REFERENCES assets(id) ON DELETE SET NULL,
  assigned_to TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  note TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS audit_log (
  id TEXT PRIMARY KEY,
  actor_id TEXT REFERENCES users(id) ON DELETE SET NULL,
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO users (id, email, name, role, password_hash) VALUES
('u-analyst', 'analyst@missionops.local', 'Maya Analyst', 'analyst', 'mission123'),
('u-operator', 'operator@missionops.local', 'Owen Operator', 'operator', 'operator123'),
('u-commander', 'commander@missionops.local', 'Casey Commander', 'commander', 'command123')
ON CONFLICT (email) DO NOTHING;

INSERT INTO assets (id, name, type, status, region, lat, lon, last_seen) VALUES
('a-001', 'Raven-1', 'drone', 'active', 'north', 38.950, -77.460, NOW()),
('a-002', 'Convoy-7', 'convoy', 'active', 'north', 38.925, -77.420, NOW()),
('a-003', 'Tower-12', 'sensor', 'watch', 'east', 38.890, -77.360, NOW()),
('a-004', 'Sentinel-3', 'drone', 'maintenance', 'east', 38.860, -77.320, NOW()),
('a-005', 'Harbor-2', 'vessel', 'active', 'south', 38.780, -77.180, NOW()),
('a-006', 'Patrol-5', 'vehicle', 'active', 'south', 38.805, -77.220, NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO events (id, type, source, region, asset_id, lat, lon, ts, payload_json) VALUES
('e-001', 'sensor_anomaly', 'perimeter-sensor', 'north', 'a-001', 38.952, -77.463, NOW() - INTERVAL '35 minutes', '{"confidence": 91, "summary": "Thermal spike near Raven-1"}'),
('e-002', 'route_deviation', 'ops-console', 'north', 'a-002', 38.928, -77.425, NOW() - INTERVAL '23 minutes', '{"confidence": 87, "summary": "Convoy-7 drifted off assigned route"}'),
('e-003', 'comms_loss', 'drone-feed', 'east', 'a-004', 38.861, -77.321, NOW() - INTERVAL '16 minutes', '{"confidence": 94, "summary": "Sentinel-3 lost uplink for 45 seconds"}'),
('e-004', 'geofence_breach', 'satellite-ingest', 'south', 'a-005', 38.782, -77.182, NOW() - INTERVAL '8 minutes', '{"confidence": 90, "summary": "Harbor-2 crossed restricted boundary"}')
ON CONFLICT (id) DO NOTHING;

INSERT INTO alerts (id, title, severity, status, region, asset_id, assigned_to, created_at, updated_at, note) VALUES
('al-001', 'Sensor Anomaly', 'high', 'open', 'north', 'a-001', NULL, NOW() - INTERVAL '34 minutes', NOW() - INTERVAL '34 minutes', 'Auto-generated from perimeter-sensor for Raven-1.'),
('al-002', 'Route Deviation', 'medium', 'acknowledged', 'north', 'a-002', 'u-operator', NOW() - INTERVAL '22 minutes', NOW() - INTERVAL '15 minutes', 'Operator is reviewing route exception.'),
('al-003', 'Comms Loss', 'high', 'open', 'east', 'a-004', NULL, NOW() - INTERVAL '15 minutes', NOW() - INTERVAL '15 minutes', 'Escalate if uplink loss exceeds 2 minutes.'),
('al-004', 'Geofence Breach', 'critical', 'escalated', 'south', 'a-005', 'u-commander', NOW() - INTERVAL '7 minutes', NOW() - INTERVAL '4 minutes', 'Commander notified for immediate review.')
ON CONFLICT (id) DO NOTHING;
