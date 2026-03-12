# Architecture Notes

## Goal
Build a mini common operating picture dashboard that looks like an operational decision-support product instead of a generic CRUD app.

## Core flows
1. Seed the database with users, assets, alerts, and historical events.
2. Start a simulator that creates fresh events every 5 seconds.
3. Query operational data from the frontend using filters for region, alert status, severity, and time window.
4. Allow operators and commanders to take action on alerts while analysts stay view-only.
5. Record alert actions in an audit log.

## Entities
- **User**: authenticated person with a role.
- **Asset**: tracked entity on the map.
- **Event**: raw operational data point.
- **Alert**: actionable item derived from operational activity.
- **AuditLog**: immutable history of decisions.

## Security choices
- Bearer token authentication
- Backend-enforced role checks
- Separate analyst/operator/commander privileges

## Tradeoffs
- Uses a lightweight pseudo-map instead of full GIS tiles to keep the MVP fast to run.
- Uses polling every 5 seconds instead of WebSockets for simplicity.
- Uses seeded demo credentials so recruiters can run the app immediately.
