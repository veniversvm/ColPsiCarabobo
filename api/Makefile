.PHONY: test test-unit test-repo test-integration test-security test-all test-race test-bench coverage coverage-html

# ─── Test DB ─────────────────────────────────────────────────────────────
TEST_DB_DSN ?= host=localhost port=5433 user=postgres password=postgres dbname=colpsi_test sslmode=disable

# ─── Unit Tests (no DB, fast) ────────────────────────────────────────────
test-unit:
	@echo "▸ Running unit tests..."
	go test -count=1 -short ./internal/utils/... ./internal/request_structs/... ./internal/middleware/... ./internal/service/... ./pkg/job/...

# ─── Repository Tests (real PostgreSQL) ──────────────────────────────────
test-repo:
	@echo "▸ Running repository tests..."
	TEST_DB_DSN="$(TEST_DB_DSN)" go test -count=1 ./internal/repository/...

# ─── Integration Tests (full stack: DB + Fiber + JWT) ────────────────────
test-integration:
	@echo "▸ Running integration tests..."
	TEST_DB_DSN="$(TEST_DB_DSN)" go test -count=1 -p 1 ./internal/integration/...

# ─── Security Tests (subset of integration) ──────────────────────────────
test-security:
	@echo "▸ Running security E2E tests..."
	TEST_DB_DSN="$(TEST_DB_DSN)" go test -count=1 -p 1 -run "TestSecurity" ./internal/integration/...

# ─── All Tests (serial to avoid DB race) ─────────────────────────────────
test-all:
	@echo "▸ Running full test suite (serial)..."
	TEST_DB_DSN="$(TEST_DB_DSN)" go test -count=1 -p 1 ./...

# ─── Race Detector ───────────────────────────────────────────────────────
test-race:
	@echo "▸ Running tests with race detector..."
	TEST_DB_DSN="$(TEST_DB_DSN)" go test -count=1 -race -p 1 ./...

# ─── Benchmarks ──────────────────────────────────────────────────────────
test-bench:
	@echo "▸ Running benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./internal/utils/... ./internal/service/...

# ─── Coverage Report ─────────────────────────────────────────────────────
coverage:
	@echo "▸ Generating coverage report..."
	TEST_DB_DSN="$(TEST_DB_DSN)" go test -count=1 -p 1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "▸ Summary:"
	@go tool cover -func=coverage.out | tail -1

# ─── Coverage HTML ───────────────────────────────────────────────────────
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "▸ HTML report: coverage.html"
