# Test Makefile Commands - Quick Reference

## New Commands Added

### 1. `make test-setup`
**Purpose**: Complete test environment setup - run this FIRST!

**What it does**:
- ✅ Checks if PostgreSQL database is running (starts it if needed)
- ✅ Creates the `audit/` directory for test logs
- ✅ Creates `.env.test` from example if missing
- ✅ Applies latest database migrations
- ✅ Cleans up old test artifacts

**Usage**:
```bash
cd backend
make test-setup
```

**When to use**:
- First time setting up tests
- After pulling new database migrations
- When you want a fresh test environment

---

### 2. `make test-logs`
**Purpose**: Run all tests with complete logging

**What it does**:
- ✅ Runs all tests with verbose output
- ✅ Saves complete logs to `audit/test-output.txt`
- ✅ Shows test summary (passed/failed counts)
- ✅ Lists failed test names if any

**Usage**:
```bash
cd backend
make test-logs
```

**Output files**:
- `audit/test-output.txt` - Complete test output with all logs

**When to use**:
- Debugging test issues
- Reviewing database queries during tests
- Keeping a record of test runs

---

### 3. `make test-fails-log`
**Purpose**: Run tests and save ONLY failure details

**What it does**:
- ✅ Runs all tests
- ✅ Filters and saves only failed test output
- ✅ Saves to `audit/test-failures.txt`
- ✅ Shows failure summary

**Usage**:
```bash
cd backend
make test-fails-log
```

**Output files**:
- `audit/test-failures.txt` - Only failed tests with error messages

**When to use**:
- Quick check for test failures
- CI/CD pipelines (easier to parse)
- When you only care about what broke

---

## Existing Commands (Still Available)

### `make test`
Quick test run with summary (original command)
```bash
make test
```

### `make test-summary`
Run tests, show only pass/fail status
```bash
make test-summary
```

### `make test-one TEST=<name>`
Run a single specific test
```bash
make test-one TEST=TestUser_CleanupUserData
make test-one TEST=TestCategory_Create/Success_ValidData
```

### `make test-failed`
Re-run only the tests that failed last time
```bash
make test-failed
```

### `make test-coverage`
Run tests with code coverage report
```bash
make test-coverage
# Opens coverage.html in browser
```

### `make test-db-migrate`
Apply latest database migrations
```bash
make test-db-migrate
```

### `make test-docker`
Run tests in isolated container
```bash
make test-docker
```

### `make test-clean`
Clean up test artifacts
```bash
make test-clean
```

---

## Typical Workflow

### First Time Setup
```bash
cd backend
make test-setup      # Setup everything
make test-logs       # Run tests with full logs
```

### Daily Development
```bash
make test            # Quick test run
# or
make test-fails-log  # See only failures
```

### Debugging a Failure
```bash
make test-logs                              # Full logs
cat audit/test-output.txt | grep -A 20 FAIL  # Find failure details
make test-one TEST=TestThatFailed           # Run just that test
```

### Before Committing
```bash
make test-setup      # Fresh environment
make test-coverage   # Check coverage
# Review coverage.html
```

---

## File Locations

All test artifacts are saved to the `audit/` directory:

- `audit/test-output.txt` - Full test logs (from `make test-logs`)
- `audit/test-failures.txt` - Only failures (from `make test-fails-log`)
- `audit/final-test-results.txt` - Historical test results
- `coverage.out` - Coverage data
- `coverage.html` - Coverage report (viewable in browser)

---

## Tips

1. **Always run `make test-setup` after**:
   - Pulling new code
   - Changing database schema
   - Modifying migrations

2. **Use `make test-fails-log` for CI/CD**:
   - Faster than full logs
   - Easier to parse failures
   - Smaller log files

3. **Use `make test-logs` for debugging**:
   - See all database queries
   - View GORM operations
   - Understand test flow

4. **Check the help anytime**:
   ```bash
   make help
   ```

