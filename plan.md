### Phase 1: Project Setup & Database (1 hour)
- [x] Initialize Go module
- [x] Set up project structure
- [x] Create MySQL schema
- [x] Set up database models (sqlx with DB/API separation)
- [x] Initialize React + TypeScript project
- [x] Set up Docker Compose for local dev
- [x] Follow Effective Go guidelines (naming, formatting, idioms)

### Phase 2: Backend Core (2 hours)
- [x] Implement database models (DB and API separation)
- [x] Create repositories with sqlx (optimized queries, prepared statements)
- [x] Implement crawler logic (HTML parsing, link checking)
- [x] Create worker pool for background processing
- [x] Basic error handling
- [x] Connection pooling optimization

### Phase 3: Backend API (1.5 hours)
- [x] Set up Gin router
- [x] Implement JWT authentication
- [x] Create URL management endpoints (POST/GET with pagination/sorting)
- [x] Create job control endpoints (start/stop/status)
- [x] Create results retrieval endpoints
- [x] Add CORS middleware
- [x] Add request validation

### Phase 4: Backend Testing (1 hour)
- [ ] Unit tests for crawler logic
- [ ] Unit tests for parser functions
- [ ] Integration tests for API endpoints (happy paths)
- [ ] Test authentication flow

### Phase 5: Frontend (2 hours)
- [x] Set up API client with auth
- [x] Create URL management UI
- [x] Create results table (paginated, sortable)
- [x] Implement job controls (start/stop buttons)
- [x] Add status polling (real-time updates)
- [x] Add status badges (color-coded visual indicators)
- [x] Add pagination controls
- [x] Add sortable column headers
- [x] Basic styling

### Phase 6: Integration & Polish (0.5 hours)
- [ ] End-to-end testing
- [ ] Error handling improvements
- [x] Documentation (README)
- [x] Code follows Effective Go guidelines (https://go.dev/doc/effective_go)
- [ ] Final commit cleanup

## Code Quality Checklist
- [x] Go code follows Effective Go guidelines (naming, formatting, idioms)
- [x] All queries use prepared statements (sqlx)
- [x] Explicit column selection (no SELECT *)
- [x] SQL injection prevention (whitelist for sort columns)
- [x] Connection pooling configured
- [x] Error handling at all layers
- [x] Separate DB and API models