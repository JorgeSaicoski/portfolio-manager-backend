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

.PHONY: help test test-summary test-one test-failed test-coverage test-docker test-clean test-db-migrate test-setup test-logs test-fails-log

help:
	@echo "Available targets:"
	@echo "  test             - Run all tests with summary (requires dev database running)"
	@echo "  test-setup       - Setup test environment (DB, migrations, cleanup)"
	@echo "  test-logs        - Run tests and save complete logs to audit/test-output.txt"
	@echo "  test-fails-log   - Run tests and save only failures to audit/test-failures.txt"
	@echo "  test-db-migrate  - Ensure database has latest migrations before running tests"
	@echo "  test-summary     - Run tests and show only summary"
	@echo "  test-one         - Run a specific test (usage: make test-one TEST=TestCategory_Create)"
	@echo "  test-failed      - Re-run only failed tests from last run"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-docker      - Run tests in containers (isolated environment)"
	@echo "  test-clean       - Clean up test artifacts"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - Dev database must be running: podman compose up -d portfolio-postgres"
	@echo "  - Create .env.test file (see .env.test.example)"
	@echo "  - Run test-setup to initialize test environment"
	@echo ""
	@echo "Examples:"
	@echo "  make test-setup              # Setup test environment"
	@echo "  make test-logs               # Run tests with full logging"
	@echo "  make test-fails-log          # Show only failures"
	@echo "  make test-one TEST=TestCategory_Create"
	@echo "  make test-one TEST=TestCategory_Update/NotFound_InvalidID"

test-db-migrate:
	@echo "Ensuring database has latest migrations..."
	@cd .. && podman compose restart portfolio-backend
	@sleep 3
	@echo "✓ Migrations applied (backend restarted)"

test-setup:
	@echo "========================================="
	@echo "    SETTING UP TEST ENVIRONMENT"
	@echo "========================================="
	@echo ""
	@echo "Step 1: Checking if database is running..."
	@cd .. && podman compose ps portfolio-postgres | grep -q "Up" || \
		(echo "✗ Database not running. Starting it..." && podman compose up -d portfolio-postgres && sleep 5)
	@echo "✓ Database is running"
	@echo ""
	@echo "Step 2: Creating audit directory for logs..."
	@mkdir -p audit
	@echo "✓ Audit directory ready"
	@echo ""
	@echo "Step 3: Checking .env.test configuration..."
	@if [ ! -f ../.env.test ]; then \
		echo "⚠ Warning: .env.test not found."; \
		echo "  Creating from .env.test.example..."; \
		if [ -f ../.env.test.example ]; then \
			cp ../.env.test.example ../.env.test; \
			echo "✓ Created .env.test from example"; \
		else \
			echo "✗ .env.test.example not found. Using environment defaults."; \
		fi \
	else \
		echo "✓ .env.test exists"; \
	fi
	@echo ""
	@echo "Step 4: Applying database migrations..."
	@cd .. && podman compose restart portfolio-backend
	@sleep 3
	@echo "✓ Migrations applied"
	@echo ""
	@echo "Step 5: Cleaning previous test artifacts..."
	@rm -f audit/test-*.txt coverage.out coverage.html /tmp/test-output.txt
	@echo "✓ Test artifacts cleaned"
	@echo ""
	@echo "========================================="
	@echo "  ✓ TEST ENVIRONMENT READY"
	@echo "========================================="
	@echo ""
	@echo "You can now run:"
	@echo "  make test           - Run tests with summary"
	@echo "  make test-logs      - Run tests with full logs"
	@echo "  make test-fails-log - Show only test failures"

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
	go test ./cmd/test/... -v -coverprofile=coverage.out || true
	go tool cover -html=coverage.out -o coverage.html
	@echo "\nCoverage report generated: coverage.html"
	@go tool cover -func=coverage.out
	@echo ""
	@echo "========================================="
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total Coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE >= 85" | bc -l) -eq 1 ]; then \
		echo "✅ Coverage target met (≥85%)"; \
		echo "========================================="; \
		exit 0; \
	else \
		echo "❌ Coverage below target (<85%)"; \
		echo "========================================="; \
		exit 1; \
	fi

test-docker:
	@echo "Running tests in isolated Podman environment..."
	podman compose -f ../docker-compose.test.yml run --rm portfolio-backend-test

test-clean:
	@echo "Cleaning up test artifacts..."
	rm -f coverage.out coverage.html
	rm -f audit/test-*.txt /tmp/test-output.txt

test-logs:
	@echo "========================================="
	@echo "  Running tests with full logging..."
	@echo "========================================="
	@mkdir -p audit
	@if [ ! -f ../.env.test ]; then \
		echo "Warning: .env.test not found. Using environment defaults."; \
	fi
	@echo ""
	@echo "Running tests (output saved to audit/test-output.txt)..."
	@go test ./cmd/test/... -v 2>&1 | tee audit/test-output.txt; \
	EXIT_CODE=$${PIPESTATUS[0]}; \
	echo ""; \
	echo "========================================="; \
	echo "           TEST SUMMARY"; \
	echo "========================================="; \
	PASSED=$$(grep -c "^--- PASS:" audit/test-output.txt 2>/dev/null || echo "0"); \
	FAILED=$$(grep -c "^--- FAIL:" audit/test-output.txt 2>/dev/null || echo "0"); \
	TOTAL=$$((PASSED + FAILED)); \
	echo "Total Tests: $$TOTAL"; \
	echo "Passed:      $$PASSED ✓"; \
	echo "Failed:      $$FAILED ✗"; \
	echo "========================================="; \
	echo ""; \
	echo "Full logs saved to: audit/test-output.txt"; \
	if [ $$FAILED -gt 0 ]; then \
		echo ""; \
		echo "Failed Tests:"; \
		grep "^--- FAIL:" audit/test-output.txt | sed 's/--- FAIL: /  ✗ /' || true; \
		echo ""; \
		echo "Run 'make test-fails-log' to see detailed failure logs."; \
	fi; \
	exit $$EXIT_CODE

test-fails-log:
	@echo "========================================="
	@echo "  Extracting test failures..."
	@echo "========================================="
	@mkdir -p audit
	@if [ ! -f ../.env.test ]; then \
		echo "Warning: .env.test not found. Using environment defaults."; \
	fi
	@echo ""
	@echo "Running tests and filtering failures..."
	@go test ./cmd/test/... -v 2>&1 | tee /tmp/test-full-output.txt | \
		grep -A 50 "^--- FAIL:" > audit/test-failures.txt 2>/dev/null || true; \
	EXIT_CODE=$${PIPESTATUS[0]}; \
	echo ""; \
	if [ -s audit/test-failures.txt ]; then \
		echo "========================================="; \
		echo "           FAILED TESTS"; \
		echo "========================================="; \
		cat audit/test-failures.txt; \
		echo ""; \
		echo "========================================="; \
		FAILED=$$(grep -c "^--- FAIL:" audit/test-failures.txt 2>/dev/null || echo "0"); \
		echo "Total Failures: $$FAILED"; \
		echo "========================================="; \
		echo ""; \
		echo "Detailed failure logs saved to: audit/test-failures.txt"; \
	else \
		echo "========================================="; \
		echo "  ✓ ALL TESTS PASSED!"; \
		echo "========================================="; \
		echo "" > audit/test-failures.txt; \
		echo "No failures to log."; \
	fi; \
	rm -f /tmp/test-full-output.txt; \
	exit $$EXIT_CODE
