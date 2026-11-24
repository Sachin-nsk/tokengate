# 0006 — Counterfactual baseline cost and shadow evaluation

Status: Accepted · 2026-11-23

## Context
"We cut the bill by X %" is only believable if traffic changes are excluded and quality is shown not to drop. Comparing last month to this month mixes in traffic growth and feature changes.

## Decision
- Each event stores `cost_usd` (what was paid) and `baseline_cost_usd`: the same request re-priced on the feature's `baseline_model` with no prompt caching and no response cache. Savings = Σ baseline − Σ actual.
- An observe-only baseline period (all features in `passthrough`) runs before any lever is enabled.
- A per-feature sample of optimised requests is replayed on the baseline model by the shadow judge, which records win / tie / loss. A route is promoted only when win-or-tie ≥ 95 % over a minimum sample, and rolled back automatically when it drops below.
- Shadow evaluation cost is tracked separately and shown on the dashboard.

## Alternatives
- **Period-over-period comparison:** simple, but confounded by traffic changes.
- **Offline eval set only:** cheaper, but does not reflect live traffic drift.

## Consequences
- Baseline cost reuses the served model's token counts; tokenizers can differ between models. The case study states this limitation.
- Shadow evaluation spends real money; sample rate and budget are capped per feature.
