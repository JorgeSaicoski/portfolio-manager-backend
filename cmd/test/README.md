# Backend Automated Tests

This directory contains comprehensive automated tests for the Portfolio Manager backend API.

## Overview

- **Total Tests**: 116 comprehensive test cases
- **Coverage**: All CRUD endpoints for Portfolio, Category, Project, and Section resources
- **Framework**: Go testing with testify assertions
- **Database**: PostgreSQL (shared with dev environment)
- **Configuration**: `.env.test` file for environment settings
- **Container Runtime**: Podman (preferred) or Docker

## Why Podman?

We use **Podman** as our primary container runtime for several important reasons:

- **Security**: Rootless containers by default - no root daemon required
- **Freedom**: 100% free and open source, no licensing restrictions
- **Simplicity**: No background daemon needed, simpler architecture
- **Compatibility**: Drop-in replacement for Docker with identical CLI
- **Modern**: Built with security and best practices from the ground up

All commands work with both Podman and Docker, but we recommend Podman for local development.

## Test Files

| File | Tests | Description |
|------|-------|-------------|
| `setup_test.go` | - | Test infrastructure and database setup |
| `fixtures.go` | - | Test data factories for creating test entities |
| `helpers.go` | - | HTTP request helpers and assertion utilities |
| `health_test.go` | 6 | Health endpoint tests |
| `portfolio_test.go` | 24 | Portfolio CRUD operations |
| `category_test.go` | 26 | Category CRUD operations |
| `project_test.go` | 29 | Project CRUD operations with skills/images |
| `section_test.go` | 31 | Section CRUD operations with types |

## Configuration

Tests use a `.env.test` file in the project root for configuration. This file should contain:

```env
# Database Configuration (Shared with Dev)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=portfolio_db
DB_USER=portfolio_user
DB_PASSWORD=portfolio_pass
DB_SSLMODE=disable
DB_TIMEZONE=UTC

# Test Server Configuration
PORT=8888
JWT_SECRET=test-jwt-secret-key-for-testing-only

# Testing Mode
TESTING_MODE=true
```

**Important**: Copy `.env.test.example` to `.env.test` and configure as needed.

## Running Tests

### Prerequisites

```bash
# 1. Start the dev database (tests use shared dev database)
podman compose up -d portfolio-postgres

# 2. Ensure .env.test exists (see Configuration section above)
cp .env.test.example .env.test
```

### Local Testing

```bash
# Run all tests
make test

# Run with coverage report
make test-coverage

# Clean up test artifacts
make test-clean
```

### Container Testing (Isolated Environment)

```bash
# Run tests in isolated container environment
# (Uses separate test database, not shared dev database)
make test-docker

# Note: Uses Podman by default, but works with Docker too
# To use Docker instead: Replace 'podman' with 'docker' in Makefile
```

## Test Structure

Each test file follows this pattern:

```go
func TestResource_Operation(t *testing.T) {
    t.Run("Success_Case", func(t *testing.T) {
        // Arrange
        cleanDatabase(testDB.DB)
        fixture := CreateTestFixture(testDB.DB, userID)

        // Act
        resp := MakeRequest(t, "GET", "/api/resource/123", nil, token)

        // Assert
        AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
            assert.Equal(t, expected, body["field"])
        })

        // Cleanup
        cleanDatabase(testDB.DB)
    })
}
```

## Environment Variables

Tests load configuration from `.env.test` file in the project root. The following variables are used:

### Database Connection
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_NAME` - Database name (shared with dev: portfolio_db)
- `DB_USER` - Database user (shared with dev: portfolio_user)
- `DB_PASSWORD` - Database password (shared with dev: portfolio_pass)
- `DB_SSLMODE` - SSL mode (default: disable)
- `DB_TIMEZONE` - Timezone (default: UTC)

### Test Server
- `PORT` - Test server port (default: 8888, different from dev to avoid conflicts)
- `JWT_SECRET` - JWT secret for testing
- `TESTING_MODE` - Must be "true" to bypass auth service
- `GIN_MODE` - Set to "test" for minimal output
- `LOG_LEVEL` - Set to "error" to reduce test noise

## Test Coverage Areas

### ✅ Health Endpoints
- Basic health check
- Database connectivity
- Response format validation

### ✅ Portfolio Endpoints
- Get own portfolios (paginated)
- Create portfolio
- Get portfolio by ID
- Update portfolio
- Delete portfolio
- Get public portfolio

### ✅ Category Endpoints
- Get own categories (paginated)
- Create category
- Get category by ID
- Update category
- Delete category (with cascade)
- Get public category
- Get categories by portfolio

### ✅ Project Endpoints
- Get own projects (paginated)
- Create project (with images, skills)
- Get project by ID
- Update project
- Delete project
- Get public project
- Get projects by category

### ✅ Section Endpoints
- Get own sections (paginated)
- Create section (different types)
- Get section by ID
- Update section
- Delete section
- Get public section
- Get sections by portfolio
- Filter by type

## CI/CD Integration

Tests run automatically on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`

GitHub Actions workflow:
- Sets up PostgreSQL service
- Installs Go dependencies
- Runs tests with race detection
- Generates coverage reports
- Comments coverage on PRs

## Notes

- **Shared Database**: Tests use the same dev database (portfolio_db) for simplicity in local development
- **Database Cleanup**: All tests clean database tables before and after execution for isolation
- **Testing Mode**: `TESTING_MODE=true` bypasses the auth service and uses a test user (ID: 123)
- **Auth Tokens**: Test token "test-token-123" is accepted in testing mode
- **Test Server Port**: Runs on port 8888 (configured in .env.test) to avoid conflict with dev server
- **Coverage Reports**: Generated in `coverage.html` and `coverage.out`
- **Podman**: Used by default for better security (rootless containers)
- **Container Tests**: `make test-docker` uses isolated test database (not shared)

## Future Improvements

- [ ] Integrate real auth token generation
- [ ] Add integration tests with auth service
- [ ] Add benchmark tests
- [ ] Add mutation testing
- [ ] Add API contract testing
