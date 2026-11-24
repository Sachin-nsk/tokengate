# 0004 — Events via Kafka into ClickHouse

Status: Accepted · 2026-11-23

## Context
Every request produces one wide event used for cost analytics, dashboards and the shadow judge. Writes must never slow a request, and analytics must stay fast at millions of rows.

## Decision
- The gateway produces versioned JSON events to Kafka topic `tokengate.llm_events.v1` (key: `request_id`).
- ClickHouse ingests them with a Kafka engine table and a materialized view into a `MergeTree` table; rollups are `AggregatingMergeTree` materialized views.
- The shadow judge consumes the same topic and publishes `tokengate.shadow_scores.v1`.
- Schema changes are new numbered migrations; event schema changes bump the topic version.

## Alternatives
- **Direct batched inserts from Go into ClickHouse:** simpler, but couples request handling to ClickHouse health and gives no replay or second consumer.
- **PostgreSQL / TimescaleDB:** familiar, but weaker for wide, append-only, high-volume analytics.
- **Prometheus only:** good for gateway health, wrong tool for per-request cost and high-cardinality dimensions.

## Consequences
- More moving parts in Compose (Kafka in KRaft mode, no ZooKeeper).
- Events can be replayed into a fresh ClickHouse after a schema change.
- Delivery is at-least-once; ClickHouse tables use `ReplacingMergeTree` or dedup on `request_id` where exact counts matter.
