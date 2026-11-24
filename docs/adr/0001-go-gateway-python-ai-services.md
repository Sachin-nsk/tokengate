# 0001 — Go gateway, Python AI services

Status: Accepted · 2026-11-23

## Context
The gateway sits in front of every LLM call, so its added latency and memory matter. The AI parts (judging, training a router, generating traffic) need the Python ML and LLM ecosystem. The author's background is Java/Spring; a sister project (Steadyrun) already shows Java + Python.

## Decision
Gateway in **Go 1.27**. AI services in **Python 3.14** (chosen 2026-11-24 for mature wheel support). Latest stable versions, pinned exactly.

## Alternatives
- **Spring Boot gateway:** familiar and productive, but higher memory and startup cost, and a harder sub-2 ms overhead claim. Duplicates the Java story Steadyrun already tells.
- **All Python (FastAPI / LiteLLM-style):** fastest to build, but the GIL and per-request overhead make the latency target hard.
- **Envoy / Kong with plugins:** industrial, but the interesting logic would live in plugin glue rather than code the author owns.

## Consequences
- Two toolchains and two CI jobs.
- Contracts between them are explicit: Kafka event schemas and exported model weights (ADR 0002).
- Go is a new language for the author; week 1 includes ramp-up time.
