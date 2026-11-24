# 0005 — Grafana dashboards as code, no custom frontend

Status: Accepted · 2026-11-23

## Context
The project needs dashboards for cost, savings, caching, routing, quality and gateway health. A custom web UI would take weeks and add little to a backend and infrastructure showcase.

## Decision
Grafana is the only UI. Datasources (ClickHouse, Prometheus, Tempo), dashboards (JSON) and alert rules are provisioned from `deploy/grafana/` at startup. Feature policies are managed as YAML in the repo and through a small admin API, not a UI.

Planned dashboards: Overview (actual vs baseline, savings), Features, Caching, Routing, Quality, Gateway health, Request explorer.

## Alternatives
- **Custom Next.js / React dashboard:** more control over the before/after story, but a large frontend effort outside the project's focus.
- **Grafana with click-ops:** fast to start, not reproducible or reviewable.

## Consequences
- Visual design is limited to what Grafana panels offer.
- Dashboard changes are reviewed as JSON diffs.
