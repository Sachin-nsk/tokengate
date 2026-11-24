# 0003 — Transparent Anthropic Messages API passthrough

Status: Accepted · 2026-11-23

## Context
Apps should adopt Tokengate by changing only the base URL of the official Anthropic SDK. The API evolves quickly (new fields, beta headers, new content block types). A gateway that decodes requests into typed structs and re-encodes them silently drops anything it does not know.

## Decision
- The gateway exposes `POST /v1/messages` (and `count_tokens`) with the same shape as the Anthropic API.
- It **forwards raw HTTP**: the body is parsed as generic JSON only for the fields it reads or rewrites (`model`, `cache_control`, `output_config.effort`), all other fields and headers pass through untouched.
- Streaming responses are forwarded event by event without buffering; the gateway reads `message_start` and `message_delta` events to collect `usage` and `stop_reason`.
- Python services, which are ordinary clients, use the official `anthropic` SDK.

## Alternatives
- **Official Go SDK for upstream calls:** typed and convenient, but re-serialising requests through SDK types can drop or alter unknown fields; acceptable for clients, not for a transparent proxy.
- **OpenAI-compatible API surface:** broader reach, but lossy for Claude-specific features (prompt caching, effort, thinking). Possible later as an extension adapter.

## Consequences
- The gateway must keep up with how `usage` is reported (including cache and thinking tokens); contract tests against recorded real responses guard this.
- Model rewrites must respect model-specific rules (e.g. features or parameters a smaller model rejects); the policy engine validates routes against a capability table.
