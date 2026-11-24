# 0002 — No network hop to Python on the request path; fail open

Status: Accepted · 2026-11-23

## Context
Routing decisions use learned models, which are easiest to build in Python. Calling a Python service on every request would add latency and a new failure point. The gateway also depends on Redis and Kafka, which can fail.

## Decision
- Models are **trained in Python and served in Go**: the router is a small model (logistic regression or gradient-boosted trees) exported as JSON weights and evaluated in-process.
- Python services run **asynchronously** (shadow judge reads Kafka; training is offline).
- Every dependency **fails open**: Redis has a 5 ms timeout and a miss is assumed; Kafka uses a bounded buffer that drops events with a counted metric when full; the policy store keeps the last good config. The client request always reaches the provider.

## Alternatives
- **Python router over gRPC per request:** flexible, but adds a hop and a dependency to every call.
- **ONNX runtime in Go:** supports any model, but needs cgo and native libraries, which complicates builds (especially on Windows) for little gain at this model size.
- **Fail closed:** safer for budgets, but turns any cache outage into a full outage.

## Consequences
- Router models are limited to what can be evaluated simply in Go.
- Budgets can be briefly exceeded while Redis is down; the overspend is visible in ClickHouse.
- The optional semantic cache (extension) is the one exception, with a 30 ms timeout and fail-open, and needs its own ADR if added.
