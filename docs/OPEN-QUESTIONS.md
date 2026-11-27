# Open questions

None open.

## Resolved
- Name: Tokengate (2026-11-24).
- Default display currency: INR (2026-11-24); USD stays selectable.
- GitHub: public repo under `sachin-nsk`.
- License: Apache-2.0.
- Python: 3.14 (ADR 0001).
- Local models: Ollama upstream (ADR 0007).
- Currency: USD storage, configurable display, per-day FX history (ADR 0008).
- Paid API runs: **none** (2026-11-27, supersedes the earlier USD 30 budget). All development, CI, load tests and demos use the fake upstream or Ollama. The case study makes no Claude savings or quality claims; S2 and S3 stay unmeasured until a paid run is agreed.
