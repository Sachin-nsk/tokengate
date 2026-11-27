# Progress

## Week 0 — Repo and design
- [x] Repo created, plan, ADRs 0001–0006
- [x] Plan reviewed by the user

## Week 1 — Toolchain, CI, local stack
- [x] Install Go 1.27 and start Docker Desktop (local Go is blocked by Windows Application Control; builds run in Docker, see docs/DEV.md)
- [x] Go module + Python project skeletons with lint and test passing
- [x] GitHub Actions: lint, test, build, image scan (written; first run happens after push)
- [x] Compose: Kafka (KRaft), ClickHouse, Redis, Prometheus, Grafana, Tempo (all healthy, Grafana datasources connect)
- [x] Fake upstream (Anthropic-compatible mock)

## Week 2 — Passthrough proxy
- [ ] Non-streaming and SSE streaming passthrough
- [ ] API keys (hashed), feature tagging
- [ ] Usage extraction, Prometheus metrics, OTel traces, graceful shutdown

## Week 3 — Metering pipeline
- [ ] Events → Kafka → ClickHouse
- [ ] Cost and baseline cost, price table
- [ ] Grafana Overview and Features dashboards
- [ ] Traffic generator, baseline run

## Week 4 — Caching
- [ ] Prompt-cache injector, cache-breaker detection
- [ ] Exact response cache with SSE replay
- [ ] Fail-open and chaos tests

## Week 5 — Policies and routing rules
- [ ] YAML policies with hot reload
- [ ] Rules router + effort
- [ ] Budgets, rate limits, admin API

## Week 6 — Quality
- [ ] Shadow judge
- [ ] Quality dashboard, automatic promote / rollback

## Week 7 — Learned routing
- [ ] Router training (Python) and serving (Go)
- [ ] Cascade with escalation, routing dashboard

## Week 8 — Proof
- [ ] k6 load tests (S1)
- [ ] Case study, README
