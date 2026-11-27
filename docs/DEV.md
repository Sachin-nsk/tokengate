# Development

Windows Application Control on the author's machine blocks unsigned toolchain binaries outside Program Files, so Go and Python tooling run in Docker. The same commands work on any machine with Docker.

```powershell
# Go: format, vet, test, lint
docker run --rm -v "${PWD}/gateway:/src" -w /src golang:1.27.4 sh -c "gofmt -l . && go vet ./... && go test -race ./..."
docker run --rm -v "${PWD}/gateway:/src" -w /src golangci/golangci-lint:v2.15.0 golangci-lint run

# Python: lint, types, tests
docker run --rm -v "${PWD}/ai:/app" -e UV_PROJECT_ENVIRONMENT=/tmp/venv -w /app ghcr.io/astral-sh/uv:python3.14-bookworm sh -c "uv run ruff check . && uv run ruff format --check . && uv run mypy src tests && uv run pytest -q"

# Local stack (gateway, fake upstream, Kafka, ClickHouse, Redis, Prometheus, Tempo, Grafana)
docker compose -f deploy/docker-compose.yml up -d --build
```

Grafana: http://localhost:3000 (admin / tokengate-dev, dev only). Gateway: http://localhost:8080/healthz. Fake upstream: http://localhost:9090/v1/messages.
