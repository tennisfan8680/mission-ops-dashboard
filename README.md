# Mission Ops Dashboard

A portfolio-ready defense-tech style software engineering project inspired by common operating picture workflows: live asset tracking, alert triage, operational filtering, role-based access, and simulated event ingestion.

## What is included
- Go backend with REST API
- PostgreSQL database with seeded demo data
- React + TypeScript frontend
- Docker Compose for one-command startup
- Role-based alert actions
- Simulator that generates new events and alerts every 5 seconds

## Demo accounts
- Analyst: `analyst@missionops.local` / `mission123`
- Operator: `operator@missionops.local` / `operator123`
- Commander: `commander@missionops.local` / `command123`

## Run with Docker
```bash
docker compose up --build
```

Then open:
- Frontend: `http://localhost:5173`
- Backend health: `http://localhost:8080/health`

## Project structure
```text
mission-ops-dashboard/
  backend/      Go API + simulator
  frontend/     React UI
  infra/        Postgres initialization SQL
  docs/         Design notes
  docker-compose.yml
```

## Suggested next upgrades
- Add PostGIS + true map tiles with Leaflet
- Add WebSocket streaming instead of polling
- Replace simple token auth with hashed passwords + refresh tokens
- Add search with Elasticsearch
- Add audit log viewer in the UI
