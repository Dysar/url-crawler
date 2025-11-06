# URL Crawler

Backend (Go) + MySQL + Frontend (React/TS) to crawl a URL and extract metadata.

## Backend

### Prerequisites
- Go 1.22+
- MySQL 8.0+

### Configuration (env)
- DB_HOST (default: 127.0.0.1)
- DB_PORT (default: 3306)
- DB_USER (default: root)
- DB_PASSWORD (default: root)
- DB_NAME (default: url_crawler)
- API_PORT (default: 8080)
- JWT_SECRET (default: dev-secret-change)
- ADMIN_USERNAME (default: admin)
- ADMIN_PASSWORD (default: password)

### Run locally
```
cd backend
go build -o bin/server ./cmd/server
API_PORT=8080 ./bin/server
```

### Docker Compose (dev)
```
docker compose up --build
```

### API
- POST /api/v1/auth/login {"username":"admin","password":"password"}
- GET /health

## Database Migrations
SQL files are in `backend/migrations`. Apply using your preferred tool or client.

## Frontend
Coming next (React + Vite). All fetch calls must use `getApiUrl()`.

## Notes
- No `SELECT *` in queries; explicitly list columns.
- Separate DB records from API models in Go.
- Do not use `uuid.MustParse` anywhere.

## Scalability & Design Notes
The current app is an MVP optimized for the test scope, but the design maps cleanly to scalable crawler principles:

- Politeness: enforce request timeouts and limited concurrency now; evolve to per-host queues with throttling/backoff.
- URL Frontier: in-memory worker pool now; move to Redis-backed frontier with visibility timeouts and dead-letter queues.
- Prioritization: FIFO today; add a priority field and biased queue selection (“front queues”).
- Freshness: timestamped results now; add periodic recrawl scheduler based on age/ETag/Last-Modified.
- Robustness: retries and structured errors now; add circuit breakers per host and idempotent job handling.
- Extensibility: parser/analyzers are modular; add plug-in analyzers for new signals without core changes.
- Robots.txt: skipped in MVP; add fetch+cache and respect crawl-delay/disallow.
- Observability: logs now; add metrics (QPS, error rates, latency), tracing, and dashboards for production.


