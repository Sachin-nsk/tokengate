# Tokengate — Plan

## 1. Goal

Put a gateway in front of every LLM call, log tokens, cost and latency per feature, route easy requests to cheaper models, cache repeated prompts, and show a credible before-vs-after.

### Success criteria (all measured, all in the final case study)
S2 and S3 need paid API runs, which are not planned (see `OPEN-QUESTIONS.md`). They remain the targets, and are reported as unmeasured rather than estimated from the fake upstream.
| # | Criterion | Target |
|---|---|---|
| S1 | Gateway overhead, p99, excluding upstream | < 2 ms at 1,000 req/s (k6, fake upstream) |
| S2 | Cost reduction on the demo workload | ≥ 60 % vs counterfactual baseline |
| S3 | Quality of optimised routes | shadow win-or-tie ≥ 95 % per promoted route |
| S4 | Availability when Redis, Kafka or ClickHouse is down | 0 failed client requests (chaos test) |
| S5 | Setup | `docker compose up` gives gateway, data stack, dashboards and demo traffic |

## 2. Architecture

```
 Client app ──(base_url=Tokengate, X-Tokengate-Key, X-Tokengate-Feature)──►
 ┌─────────────────────────── Go gateway ───────────────────────────┐
 │ 1 auth (hashed API keys)      6 upstream call (raw HTTP, SSE pass) │
 │ 2 policy lookup (feature)     7 validate → escalate if needed      │
 │ 3 budget / rate limit (Redis) 8 usage → cost + baseline_cost       │
 │ 4 response cache (Redis)      9 event → Kafka (async, bounded)     │
 │ 5 prompt-cache injector,     10 shadow sample flag                 │
 │   router (rules / model)                                           │
 └───────────────────────────────────────────────────────────────────┘
          │ events (Kafka: tokengate.llm_events.v1)
          ├──► ClickHouse (Kafka engine → MergeTree → rollup MVs)
          └──► Python shadow judge ──► Kafka (tokengate.shadow_scores.v1) ──► ClickHouse
 Prometheus ◄── gateway /metrics      Tempo ◄── OTel traces
 Grafana ◄── ClickHouse + Prometheus + Tempo
```

## 3. Components

| Component | Language | Responsibility |
|---|---|---|
| `gateway` | Go | Proxy, policies, caches, router, metering, metrics, traces |
| `fake-upstream` | Go | Anthropic-compatible mock with configurable latency, usage and errors; used by CI and load tests |
| `ai/judge` | Python | Consumes sampled events, replays on the baseline model, scores with an LLM judge or exact/schema checks |
| `ai/router_training` | Python | Trains the request-difficulty model from logged features + judge labels; exports weights (JSON) for Go |
| `ai/fx` | Python | Optional daily FX rate loader into ClickHouse (ADR 0008) |
| `ai/traffic` | Python | Demo workload generator using the official SDK pointed at Tokengate |
| `deploy/` | — | Compose, ClickHouse migrations, Grafana provisioning, Prometheus rules |

## 4. Request lifecycle details

- **Feature policy** (YAML, hot-reloaded):
  ```yaml
  features:
    ticket-classifier:
      baseline_model: claude-opus-5-5
      route: { mode: rules, model: claude-haiku-5-5, effort: low }
      response_cache: { enabled: true, ttl: 24h }
      prompt_cache: auto
      shadow_sample_rate: 0.05
      budget: { amount: 20, currency: USD, per: day }   # any supported currency (ADR 0008)
    local-tagger:
      baseline_model: claude-haiku-5-5
      route: { mode: rules, model: llama3.1:8b, upstream: ollama }   # ADR 0007
      prompt_cache: off          # Ollama does not support cache_control
    support-chat:
      baseline_model: claude-opus-5-5
      route: { mode: learned, candidates: [claude-haiku-5-5, claude-sonnet-5-5, claude-opus-5-5] }
      response_cache: { enabled: false }
  ```
- **Response cache key:** SHA-256 of canonical JSON of (served model, system, messages, tools, tool_choice, output_config, max_tokens) + feature + tenant. Streaming hits are replayed as SSE.
- **Prompt-cache injector:** adds `cache_control` to the last stable block (tools → system) when the feature allows it; detects cache breakers (changing system prefix hash across requests) and reports them as a metric.
- **Router order:** effort reduction on the same model is tried before switching models (one cache namespace; caches are model-scoped). Cascade escalation triggers: schema validation failure, `stop_reason` of `refusal` or `max_tokens`, or a feature-defined check.
- **Fail-open:** Redis timeout 5 ms, cache miss assumed; Kafka producer uses a bounded in-memory buffer, drops with a counter when full; policy store keeps last good config.

## 5. Data model (ClickHouse)

`llm_events` (one row per client request):

| Column | Type | Notes |
|---|---|---|
| `event_time` | DateTime64(3) | |
| `request_id` | UUID | |
| `tenant`, `feature`, `api_key_id`, `env` | LowCardinality(String) | |
| `requested_model`, `served_model`, `effort` | LowCardinality(String) | |
| `route_reason` | LowCardinality(String) | `rules`, `learned`, `escalated`, `passthrough` |
| `response_cache` | Enum | `hit`, `miss`, `bypass` |
| `input_tokens`, `output_tokens`, `cache_read_tokens`, `cache_write_tokens` | UInt32 | from provider `usage` |
| `ttft_ms`, `total_ms`, `gateway_overhead_ms` | Float32 | |
| `status`, `stop_reason`, `error_type` | LowCardinality(String) | |
| `cost_usd`, `baseline_cost_usd` | Decimal(18,9) | |
| `price_version` | LowCardinality(String) | |
| `prompt_prefix_hash`, `request_hash` | FixedString(32) | no raw text |
| `shadow_sampled` | Bool | |

`shadow_scores`: `request_id`, `verdict` (win/tie/loss), `score`, `judge_model`, `judge_cost_usd`.
Rollups: `cost_by_feature_hourly`, `latency_by_route_hourly` (AggregatingMergeTree MVs). TTL: raw 90 days, rollups 2 years.

## 6. Cost model

`cost = Σ tokens_k × price_k(served_model)` for k ∈ {input, output, cache_read, cache_write}.
`baseline_cost` = same token counts re-priced on `baseline_model` with every cached token priced as plain input, and response-cache hits priced as a full call. (Token counts differ slightly between models; the case study states this limitation.)
Prices live in `deploy/prices.yaml`, versioned, with the source date. Each entry has a currency; non-USD entries (e.g. local-model electricity cost in INR) are converted to USD with `deploy/fx.yaml` at load time. Events always store USD; dashboards convert to the selected display currency (ADR 0008). Local-model cost defaults to 0 and is reported as an estimate (ADR 0007).

## 7. Demo workload

| Feature | Shape | Expected main lever |
|---|---|---|
| `ticket-classifier` | short input, label output, many repeats | response cache + Haiku |
| `rag-qa` | large stable context + question | prompt caching |
| `support-chat` | multi-turn, mixed difficulty | learned router |
| `summarizer` | long input, medium output | lower effort / Sonnet |
| `code-review` | hard reasoning | stays on Opus (control group) |

## 8. Milestones (8 weeks)

| Week | Deliverable |
|---|---|
| 1 | Toolchain, CI (lint, test, build, image scan), Compose stack (Kafka KRaft, ClickHouse, Redis, Prometheus, Grafana, Tempo), fake upstream |
| 2 | Passthrough proxy: streaming + non-streaming, API keys, feature tagging, usage extraction, Prometheus metrics, OTel traces, graceful shutdown |
| 3 | Events → Kafka → ClickHouse, cost + baseline cost, Grafana Overview and Features dashboards, traffic generator, baseline run |
| 4 | Prompt-cache injector + cache-breaker detection, exact response cache (incl. SSE replay), fail-open + chaos tests |
| 5 | Policy engine (YAML hot reload), rules router + effort, budgets and rate limits, admin API (OpenAPI) |
| 6 | Shadow judge, quality dashboard, automatic promote / rollback of routes |
| 7 | Learned router (train in Python, serve in Go), cascade with escalation, routing dashboard |
| 8 | k6 load tests, case study (fake-upstream and local measurements only, no Claude savings claims), README polish |

## 9. Scope

**In:** everything in §8.
**Extensions (only if re-added):** semantic cache (Python embedding service over gRPC, 30 ms budget), Batch API routing for offline features, OpenAI provider adapter, Helm chart / Kubernetes, multi-region.
**Out:** custom web frontend (Grafana only), billing/invoicing, prompt management.

## 10. Risks

| Risk | Mitigation |
|---|---|
| Gateway is a single point of failure | stateless replicas, fail-open dependencies, health checks |
| Wrong cached answers | per-feature opt-in, never for personalised features, cache key includes tenant |
| Savings overstated | counterfactual baseline + shadow scores (D6), limitations section in case study |
| Shadow evaluation and live runs cost money | sample caps, per-run budget agreed with the user, fake upstream for CI |
| Price changes | versioned price table stamped on each event |
| Local judge flatters local routes | calibrate against the Claude judge before use (ADR 0007) |
| FX moves distort before/after | pinned-rate toggle, rate version recorded (ADR 0008) |
