# 0008 — Currency: store USD, convert by configuration

Status: Accepted · 2026-11-24

## Context
Anthropic bills in USD, but users want to see and budget in INR, EUR and others. Storing converted amounts would bake a moving exchange rate into history and make every currency change a data rewrite.

## Decision
- **Canonical storage is USD.** `cost_usd` and `baseline_cost_usd` stay in USD, matching provider invoices. Tokens remain the source of truth.
- **Config** (`deploy/config.yaml`, env overrides):
  ```yaml
  currency:
    display: INR            # default for dashboards and reports
    supported: [USD, INR, EUR, GBP]
    fx:
      source: file          # file | ecb | static
      file: deploy/fx.yaml  # pinned rates, used by tests and the case study
  ```
- **FX table** `fx_rates(date, currency, per_usd, source)` lives in ClickHouse, loaded from `deploy/fx.yaml` or a daily job (Python) when `source` is `ecb`. Rates are versioned; the case study pins one.
- **History never moves.** Each event is converted at the rate of *its own date*, never today's rate. If an event cost ₹100 on Monday it still shows ₹100 on Tuesday after the dollar rises to ₹110; Tuesday's events use Tuesday's rate. Rules that make this true:
  - `fx_rates` is **append-only**: a published day's rate is never updated or backfilled. A correction is a new row with a higher `revision`, and dashboards keep the revision that was current when the day closed.
  - A date with no rate falls back to the **latest earlier** rate, never the latest overall.
  - Enforced by a test: change today's rate, assert that every past day's converted totals are unchanged.
- **Actual bills:** the converted figure is an estimate at the daily rate. What a bank or card really charged in INR includes its own rate and fees. An optional `billed_amounts` table (month, currency, amount paid) feeds a reconciliation panel showing estimate vs actual, so the true spend can be tracked too.
- **Display:** a ClickHouse dictionary joins events to the rate on the event date (rules above). Grafana has a `$currency` variable; panels multiply `cost_usd` by the rate and use the matching currency unit. A dashboard-wide "pinned rate" toggle uses a single rate for the whole range so before/after comparisons are not distorted by FX moves.
- **Inputs in other currencies** (budgets, local-model price per token) are written as `{amount, currency}`, converted to USD at load time, and the rate version is logged. Redis budget counters are kept in micro-USD.
- Money is never a float: Decimal in ClickHouse, integer micro-units in Go and Redis, `Decimal` in Python.

## Alternatives
- **Store converted cost per event:** one column, simple queries, but history changes meaning when the display currency changes, and invoices would not reconcile.
- **Single global currency setting at deploy time:** simplest, but different teams could not view the same data in their own currency.

## Consequences
- Dashboards cost one extra join and a variable; negligible at this scale.
- Grafana has no lakh/crore digit grouping for INR; amounts display in the standard grouping.
- Tax (GST/VAT) is out of scope: all figures are pre-tax provider prices.
