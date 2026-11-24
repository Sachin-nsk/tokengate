# 0007 — Local models via Ollama as an upstream

Status: Accepted · 2026-11-24

## Context
Paid API calls are the main cost of development, shadow evaluation and demo traffic. Ollama runs open models locally and exposes an Anthropic-compatible `/v1/messages` endpoint, so it fits the transparent passthrough of ADR 0003. Its compatibility is partial: prompt caching (`cache_control`), `count_tokens`, `tool_choice` and batches are not supported, token counts are approximate, and extended-thinking budgets are not enforced.

## Decision
- An **upstream** is a named entry in config: `{name, base_url, auth, wire: anthropic}`. Policies route to `model` + `upstream`; the default upstream is Anthropic. Ollama is a second upstream with the same wire format, so no request translation is needed.
- A **capability table** per upstream/model lists unsupported features. The policy engine rejects a route that needs one (for example a feature with `prompt_cache: auto` routed to Ollama) at load time, and the router strips nothing silently.
- **Price table** entries for local models carry an optional per-token cost covering electricity or hardware amortisation (default 0), in any configured currency (ADR 0008). Savings are reported both with that cost and with it at zero.
- Ollama is usable in three places: as a routing target for easy features, as the backend for the **fake-upstream-with-real-text** dev mode, and as a **cheaper shadow judge**.
- A local judge is only used for features whose promotion gate has been calibrated against the Claude judge on a labelled sample (agreement ≥ 90 %); otherwise judging stays on Claude.

## Alternatives
- **OpenAI-compatible adapter for Ollama:** needs request and response translation, which is lossy for Claude features. Kept as a possible extension.
- **LiteLLM in front of Ollama:** quick, but hides the behaviour this project is meant to demonstrate.

## Consequences
- Quality claims for local routes rest on shadow evaluation like any other route; a weak local judge could inflate them, hence the calibration rule.
- Ollama usage figures are approximate; cost for local models is an estimate and is labelled as such on dashboards.
- Local latency depends on the developer's hardware, so load-test numbers still use the fake upstream.
