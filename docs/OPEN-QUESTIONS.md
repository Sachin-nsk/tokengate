# Open questions

1. **Default display currency.** USD or INR for the demo dashboards and case study?

## Resolved
- Name: Tokengate (2026-11-24).
- GitHub: public repo under `sachin-nsk`, created after the first reviewed commit.
- License: Apache-2.0 (add `LICENSE` and `NOTICE` in the first commit).
- Python: 3.14 (ADR 0001).
- Local models: Ollama upstream (ADR 0007).
- Currency: USD storage, configurable display, per-day FX history (ADR 0008).
- Live-run budget (delegated to Claude, 2026-11-24): hard cap of **USD 30 in total**, enforced by the gateway's own budget feature: baseline run $5, final run $5, shadow evaluation $10, reserve $10. Dev and CI traffic uses the fake upstream or Ollama. No live run starts without a fresh go-ahead from the user.
