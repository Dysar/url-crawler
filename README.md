# URL Crawler

Backend (Go) + MySQL + Frontend (React/TS) to crawl a URL and extract:
HTML version, title, H1–H6 counts, internal/external links, inaccessible links, and login form presence.

## Demo

![URL Crawler Demo](assets/url-crawl.gif)

## Run it (single command)

```
docker compose up --build
```

Then open the UI at http://localhost:3000 (API at http://localhost:8080).

## Configuration

- DB_HOST (default: 127.0.0.1)  
- DB_PORT (default: 3306)  
- DB_USER (default: root)  
- DB_PASSWORD (default: root)  
- DB_NAME (default: url_crawler)  
- API_PORT (default: 8080)  
- JWT_SECRET (default: dev-secret-change)  
- ADMIN_USERNAME (default: admin)  
- ADMIN_PASSWORD (default: password)

## Architecture at a glance

- Backend: Go (Gin), MySQL, sqlx; clean layering: `api` → `service` → `repository` → `db/models`
- Frontend: React + TypeScript + Vite; simple stateful table with polling
- Worker pool: bounded channel + N workers (default 10) for concurrent crawling
- Auth: JWT (Bearer) on secured routes

## Key decisions & trade‑offs (per requirements)

- **MySQL schema (ENUM ‘queued|running|done|error|stopped’)**  
  Trade‑off: ENUM enforces valid states and keeps queries fast; requires migration when adding states.
- **Bounded worker pool (10) + buffered queue (100)**  
  Trade‑off: Simple, safe back‑pressure; avoids unbounded goroutines. Can be made configurable later.
- **Crawl accuracy rules**  
  - HTML version from doctype; default to HTML5 when unknown.  
  - Headings counted per tag (H1–H6).  
  - Links: only http/https; relative links resolved against `<base>` or request URL.  
  - Inaccessible links checked with timeouts; counted (not stored in DB).  
  - Login form: presence of `<input type="password">`.
- **Status flow “queued → running → done/error”**  
  Requirement-aligned text while keeping internal code identifiers stable.
- **CORS**  
  Explicit allow for `Authorization` header to keep browser requests reliable.

## Testing

```bash
cd backend
go test ./...                                # unit tests
E2E_TEST=1 go test -tags=e2e ./internal/crawler   # real HTTP e2e
```

Coverage targets the requirements: HTML version, title, H1–H6, internal/external links, inaccessible links, login form.

## Test URLs

- Minimal: `http://example.com`, `http://httpbin.org/html`  
- Headings: `https://www.w3.org/Style/Examples/011/firstcss.en.html`  
- Many links: `https://www.w3.org/TR/PNG/`  
- Forms: `http://httpbin.org/forms/post`, Login: `https://github.com/login`  
- Errors: `https://httpstat.us/404`, `https://httpstat.us/500`
