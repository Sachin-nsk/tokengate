# Tokengate

An LLM gateway that sits in front of every model call, measures what each feature costs, and cuts the bill without lowering answer quality.

> Status: design phase. See [docs/PLAN.md](docs/PLAN.md) and [docs/PROGRESS.md](docs/PROGRESS.md).

## What it does

- **Meters every call:** tokens, cost, latency and cache usage per feature, tenant and model.
- **Caches:** provider prompt caching injected automatically, plus an exact response cache for features that opt in.
- **Routes:** sends easy requests to a smaller model, a lower effort level, or a free local model (Ollama); escalates when the cheap answer fails validation.
- **Proves the savings:** every request records what it *would* have cost on the original model with no caching, and a shadow evaluator checks that routed answers are as good as the originals.
- **Speaks your currency:** costs are stored in USD (what the provider bills) and shown in USD, INR, EUR or any configured currency.

Apps change only their base URL and add an `X-Tokengate-Feature` header.

## Architecture

```mermaid
flowchart LR
    classDef app fill:#eef2ff,stroke:#6366f1,color:#1e1b4b
    classDef gw fill:#ecfdf5,stroke:#10b981,color:#064e3b
    classDef store fill:#fff7ed,stroke:#f97316,color:#7c2d12
    classDef ai fill:#fdf2f8,stroke:#ec4899,color:#831843
    classDef ext fill:#f1f5f9,stroke:#64748b,color:#0f172a
    classDef obs fill:#fefce8,stroke:#eab308,color:#713f12

    APP["Your apps<br/>official Anthropic SDK<br/>base_url = Tokengate"]:::app

    subgraph GW["Go gateway - request path, p99 overhead under 2 ms"]
        direction TB
        A1["Auth + feature policy"] --> A2["Budget + rate limit"]
        A2 --> A3["Response cache lookup"]
        A3 --> A4["Prompt-cache injector"]
        A4 --> A5["Router<br/>rules / learned model"]
        A5 --> A6["Upstream call<br/>SSE streamed, unbuffered"]
        A6 --> A7["Validate / escalate"]
        A7 --> A8["Cost + baseline cost"]
    end
    class GW gw

    subgraph UP["Upstream models"]
        direction TB
        ANT["Anthropic API<br/>Haiku / Sonnet / Opus"]:::ext
        OLL["Ollama<br/>local models"]:::ext
    end

    subgraph ST["State and events"]
        direction TB
        RED[("Redis<br/>exact cache, budgets")]:::store
        KAF[["Kafka<br/>llm_events.v1"]]:::store
        CH[("ClickHouse<br/>events + rollups + FX rates")]:::store
    end

    subgraph PY["Python AI services - off the request path"]
        direction TB
        JUDGE["Shadow judge"]:::ai
        TRAIN["Router training"]:::ai
        TRAF["Traffic generator"]:::ai
    end

    subgraph OBS["Observability"]
        direction TB
        PROM["Prometheus"]:::obs
        TEMPO["Tempo traces"]:::obs
        GRAF["Grafana<br/>dashboards as code<br/>currency selector"]:::obs
    end

    APP ==> A1
    A6 ==> ANT
    A6 -.-> OLL
    A3 <-.->|"fail open"| RED
    A2 <-.->|"fail open"| RED
    A8 -->|"async, bounded buffer"| KAF
    KAF --> CH
    KAF --> JUDGE
    JUDGE -->|"win / tie / loss"| KAF
    JUDGE -.->|"replay on baseline model"| ANT
    CH -->|"labels + features"| TRAIN
    TRAIN -.->|"exported weights, loaded by Go"| A5
    TRAF --> APP
    GW -.->|"/metrics, OTel"| PROM
    GW -.-> TEMPO
    CH --> GRAF
    PROM --> GRAF
    TEMPO --> GRAF
```

Solid arrows are the data path, dotted arrows are optional or best-effort. Redis, Kafka and ClickHouse can all be down and requests still reach the provider.

### Life of a request

```mermaid
sequenceDiagram
    autonumber
    participant App as Client app
    participant GW as Tokengate gateway
    participant R as Redis
    participant M as Upstream model
    participant K as Kafka
    participant CH as ClickHouse
    participant J as Shadow judge

    App->>GW: POST /v1/messages + X-Tokengate-Feature
    GW->>GW: authenticate, load feature policy
    GW->>R: budget check, response-cache lookup (5 ms timeout)
    alt cache hit and feature opted in
        R-->>GW: stored response
        GW-->>App: replay (SSE if streaming)
    else miss, Redis down or cache disabled
        GW->>GW: inject cache_control, pick model and effort
        GW->>M: forward raw request
        M-->>GW: stream tokens + usage
        GW-->>App: stream unbuffered
        opt cheap answer fails validation
            GW->>M: escalate to a stronger model
        end
    end
    GW-)K: event with cost and baseline_cost (async, never blocks)
    K-)CH: ingest
    K-)J: sampled requests
    J->>M: replay on baseline model, score
    J-)K: verdict
```

| Part | Tech |
|---|---|
| Gateway (request path) | Go |
| AI services (router training, shadow judge, traffic generator) | Python 3.14 + official `anthropic` SDK |
| Models | Anthropic API; Ollama for local, near-free tiers |
| Events | Kafka → ClickHouse |
| Cache, budgets, rate limits | Redis |
| Dashboards | Grafana, provisioned from this repo |
| Metrics and traces | Prometheus, OpenTelemetry |

## Repository layout

| Path | Contents |
|---|---|
| `gateway/` | Go gateway |
| `ai/` | Python services |
| `proto/` | gRPC contracts |
| `api/` | OpenAPI spec for the admin API |
| `deploy/` | Docker Compose, ClickHouse schema, Grafana dashboards, Prometheus rules, price and FX tables |
| `load/` | k6 load tests |
| `docs/` | Plan, progress, ADRs, runbook |

## Design decisions

See [docs/adr/](docs/adr/README.md).
