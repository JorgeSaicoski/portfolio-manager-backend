# Portfolio Manager Backend - Test Makefile
#
# Container Runtime: Podman
# We use Podman instead of Docker for security and freedom reasons:
# - Rootless containers by default (better security)
# - No daemon required (simpler architecture)
# - Drop-in replacement for Docker (same CLI)
# - Free and open source (no licensing concerns)
#
# Note: These commands work with both Podman and Docker
#
# Test Database Strategy:
# - Tests use the SHARED development database (not isolated)
# - Configuration is loaded from .env.test file in project root
# - This simplifies local development (no separate test DB needed)
# - The dev database (portfolio-postgres) must be running
# - Tests clean up data between runs using database transactions

.PHONY: help test test-summary test-one test-failed test-coverage test-docker test-clean

help:
	@echo "Available targets:"
	@echo "  test          - Run all tests with summary (requires dev database running)"
	@echo "  test-summary  - Run tests and show only summary"
	@echo "  test-one      - Run a specific test (usage: make test-one TEST=TestCategory_Create)"
	@echo "  test-failed   - Re-run only failed tests from last run"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-docker   - Run tests in containers (isolated environment)"
	@echo "  test-clean    - Clean up test artifacts"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - Dev database must be running: podman compose up -d portfolio-postgres"
	@echo "  - Create .env.test file (see .env.test.example)"
	@echo ""
	@echo "Examples:"
	@echo "  make test-one TEST=TestCategory_Create"
	@echo "  make test-one TEST=TestCategory_Update/NotFound_InvalidID"

test:
	@echo "Running tests with shared dev database..."
	@if [ ! -f ../.env.test ]; then \
		echo "Warning: .env.test not found. Using environment defaults."; \
		echo "See .env.test.example for configuration options."; \
	fi
	@go test ./cmd/test/... -v 2>&1 | tee /tmp/test-output.txt; \
	EXIT_CODE=$${PIPESTATUS[0]}; \
	echo ""; \
	echo "========================================="; \
	echo "           TEST SUMMARY"; \
	echo "========================================="; \
	PASSED=$$(grep -c "^--- PASS:" /tmp/test-output.txt || echo "0"); \
	FAILED=$$(grep -c "^--- FAIL:" /tmp/test-output.txt || echo "0"); \
	TOTAL=$$((PASSED + FAILED)); \
	echo "Total Tests: $$TOTAL"; \
	echo "Passed:      $$PASSED ✓"; \
	echo "Failed:      $$FAILED ✗"; \
	echo "========================================="; \
	if [ $$FAILED -gt 0 ]; then \
		echo ""; \
		echo "Failed Tests:"; \
		grep "^--- FAIL:" /tmp/test-output.txt | sed 's/--- FAIL: /  ✗ /' || true; \
	fi; \
	rm -f /tmp/test-output.txt; \
	exit $$EXIT_CODE

test-summary:
	@echo "Running tests (summary only)..."
	@if [ ! -f ../.env.test ]; then \
		echo "Warning: .env.test not found. Using environment defaults."; \
	fi
	@go test ./cmd/test/... 2>&1 | grep -E "^(PASS|FAIL|ok|FAIL)" || go test ./cmd/test/...

test-one:
	@if [ -z "$(TEST)" ]; then \
		echo "Error: TEST variable not set"; \
		echo "Usage: make test-one TEST=TestCategory_Create"; \
		echo "   or: make test-one TEST=TestCategory_Update/NotFound_InvalidID"; \
		exit 1; \
	fi
	@echo "Running test: $(TEST)"
	@if [ ! -f ../.env.test ]; then \
		echo "Warning: .env.test not found. Using environment defaults."; \
	fi
	@go test ./cmd/test/... -v -run "^$(TEST)$$"

test-failed:
	@if [ ! -f /tmp/test-output.txt ]; then \
		echo "No previous test run found. Run 'make test' first."; \
		exit 1; \
	fi
	@echo "Re-running failed tests..."
	@FAILED_TESTS=$$(grep "^--- FAIL:" /tmp/test-output.txt | sed 's/--- FAIL: //' | sed 's/ .*//' | tr '\n' '|' | sed 's/|$$//'); \
	if [ -z "$$FAILED_TESTS" ]; then \
		echo "No failed tests found!"; \
	else \
		echo "Failed tests: $$FAILED_TESTS"; \
		go test ./cmd/test/... -v -run "$$(echo $$FAILED_TESTS | sed 's/|/$$|/g')$$"; \
	fi

test-coverage:
	@echo "Running tests with coverage..."
	@if [ ! -f ../.env.test ]; then \
		echo "Warning: .env.test not found. Using environment defaults."; \
	fi
	go test ./cmd/test/... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "\nCoverage report generated: coverage.html"
	go tool cover -func=coverage.out

test-docker:
	@echo "Running tests in isolated Podman environment..."
	podman compose -f ../docker-compose.test.yml run --rm portfolio-backend-test

test-clean:
	@echo "Cleaning up test artifacts..."
	rm -f coverage.out coverage.html
