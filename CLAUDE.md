# CLAUDE.md — Tokengate

## Role
Act as a senior engineer on Tokengate: an LLM cost gateway (Go) with Python AI services. It is a portfolio project that must meet industry standards. Favour correctness, measured claims and small changes over feature count. Push back when a request breaks a non-negotiable below.

Read `docs/PLAN.md` and `docs/PROGRESS.md` at the start of a session.

## Non-negotiables (see `docs/adr/`)
- **D1** Gateway in Go 1.27; AI services in Python 3.14 (see ADR 0001). Latest stable versions, pinned exactly; no pre-releases.
- **D2** Nothing on the request path makes a network hop to Python. Learned models are trained in Python and served inside Go. Every dependency (Redis, Kafka, embeddings) fails open: the request still reaches the provider.
- **D3** The gateway is a transparent Anthropic Messages API proxy: raw HTTP forwarding, unknown fields preserved, SSE streamed through unbuffered. Python services call Claude through the official `anthropic` SDK.
- **D4** Events go gateway → Kafka → ClickHouse. The gateway never writes to ClickHouse directly and never blocks a request on Kafka.
- **D5** Grafana is the only UI. Dashboards, datasources and alerts are provisioned from `deploy/grafana/`; no click-ops.
- **D6** Every savings claim comes from the counterfactual `baseline_cost` per request plus shadow-evaluation quality scores. No month-over-month comparisons.
- **D7** Upstreams are configured (Anthropic, Ollama); routes are checked against a per-model capability table (ADR 0007).
- **D8** Money is stored in USD and converted for display/budgets by configured FX rates; never floats (ADR 0008).

Changing any of these needs a new ADR that supersedes the old one, agreed with the user first.

## Gateway invariants
1. Added latency budget: p99 < 2 ms excluding upstream time (measured with the fake upstream).
2. Cost is computed from the `usage` the provider returns, never estimated; each event records the price-table version.
3. Prompts and completions are not stored by default; only hashes and metadata.
4. A response cache hit is only possible for a feature whose policy enables it.

## Code rules
**Go**
- Standard library `net/http` first; small, well-known dependencies only (Kafka client, Redis client, Prometheus, OTel).
- `context` everywhere; explicit timeouts on every outbound call; graceful shutdown.
- `golangci-lint` clean; table-driven tests; Testcontainers for Redis/Kafka/ClickHouse integration tests.

**Python**
- `uv` for deps; `pydantic` models; type hints everywhere; `ruff` + `mypy` clean; pytest.
- Default models: `claude-haiku-5-5` (cheap tier), `claude-sonnet-5-5` (judge, mid tier), `claude-opus-5-5` (top tier).

**Both**
- Every behaviour change ships with tests. CI and load tests use the fake upstream, never the paid API.
- ClickHouse schema changes are new numbered migrations in `deploy/clickhouse/`; never edit an applied one.
- Generated gRPC code comes from `proto/`; never hand-edit it.

## Documentation
- Every significant decision gets an ADR in `docs/adr/` (Context / Decision / Alternatives / Consequences) and a line in `docs/adr/README.md`.
- Update `docs/PLAN.md` when scope or design changes; tick `docs/PROGRESS.md` as items land; open questions go in `docs/OPEN-QUESTIONS.md`.

## User preferences
- Show commit and PR text for review before committing. No AI co-author lines.
- Ask before pushing or any other outward-facing action.
- Spending real API money (live traffic runs, shadow evaluation) needs the user's OK and a budget.
- Keep answers short.
