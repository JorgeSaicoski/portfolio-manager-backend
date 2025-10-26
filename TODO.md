# Backend Refactoring TODO

## **1. New Image Model**

### Current Issue
Projects store images as `[]string` (array of URLs)

### Proposed Structure
Image Model (new):
- ID (primary key)
- URL (string)
- FileName (string)
- FileSize (int64)
- MimeType (string)
- Alt (string - for accessibility)
- OwnerID (string)
- Type (enum: "photo" | "image" | "icon" | "logo" | "banner" | "avatar" | "background")
- EntityID (uint - polymorphic relation)
- EntityType (string - "project", "portfolio", "section")
- IsMain (boolean)
- CreatedAt, UpdatedAt, DeletedAt

### Implementation Tasks
- [ ] Create `backend/internal/domain/models/image.go`
- [ ] Create `backend/internal/domain/repositories/image_repository.go`
- [ ] Create `backend/internal/application/handlers/image.go`
- [ ] Update Project model to use Image relationship instead of `[]string`
- [ ] Add image upload endpoint (POST `/api/images/upload`)
- [ ] Add image management endpoints (GET `/api/images`, DELETE `/api/images/:id`)
- [ ] Add migration to convert existing image URLs
- [ ] Update DTOs to handle Image objects
- [ ] Add HTTP tests for image endpoints

### Benefits
- Better image metadata management
- Centralized image storage info
- Easier to implement image optimization
- Better tracking of image usage
- Support for multiple entity types (polymorphic)

---

## **2. Security Audit**

### Areas to Check

**Authentication & Authorization:**
- ✅ Using Authentik (no custom auth service)
- ✅ Migrated to Authentik OAuth2/OIDC
- ✅ Backend validates tokens via OIDC
- ✅ Frontend uses OAuth2 Authorization Code flow with PKCE
- ✅ Prometheus endpoint protected with basic authentication
- ✅ Rate limiting on all endpoints (100 req/min per IP)
- [ ] Implement token refresh mechanism (future enhancement)

**Input Validation:**
- ✅ DTOs have validation tags
- ✅ SQL injection protection via GORM parameterized queries
- ✅ XSS protection utilities (SanitizeString, SanitizeHTML)
- ✅ Input sanitization middleware created
- ✅ File validation utilities (ValidateFileExtension, SanitizeFilename)
- ✅ Email and URL validation functions

**API Security:**
- ✅ CORS configured properly (environment-based origins)
- ✅ Request size limits implemented (10MB default, configurable)
- ✅ Rate limiting middleware (IP-based, 100 req/min)
- ✅ All protected endpoints use auth middleware
- ✅ Error messages sanitized in production mode
- ✅ Request ID tracking for debugging (X-Request-ID header)
- ✅ Security headers (CSP, X-Frame-Options, etc.)

**Database Security:**
- ✅ Connection strings use environment variables only
- ✅ Query timeout configuration (30s default)
- ✅ Database query logging (environment-based levels)
- ✅ Connection pooling with configurable limits
- ✅ Prepared statement caching enabled
- [ ] Review database user permissions (ops task - principle of least privilege)
- [ ] Verify soft delete implementation (future review)

### Implementation Tasks
- ✅ Created `backend/internal/shared/middleware/security.go`:
    - Security headers (X-Frame-Options, CSP, HSTS, etc.)
    - Request size validator (configurable max 10MB)
    - Request ID generation and tracking
    - Enhanced panic recovery with safe error responses
- ✅ Created `backend/internal/shared/middleware/rate_limit.go`:
    - IP-based rate limiting with automatic cleanup
    - Configurable rate (100 req/min) and window (60s)
- ✅ Created `backend/internal/shared/middleware/sanitize.go`:
    - XSS protection utilities (SanitizeString, SanitizeHTML)
    - File validation and sanitization
    - Email and URL validation
    - Filename sanitization (path traversal prevention)
- ✅ Created `backend/internal/shared/errors/handler.go`:
    - Safe error responses for clients
    - Production mode error sanitization
    - Request ID in error responses
- ✅ Added Prometheus endpoint authentication (basic auth via env vars)
- ✅ Created comprehensive `SECURITY.md` documentation
- ✅ Updated `.gitignore` for sensitive files (tokens, keys, credentials)
- ✅ Integrated all security middleware in `server.go`
- ✅ Updated `.env.example` with security configuration variables

### Security Configuration

All security features are configurable via environment variables in `.env`:

```env
# CORS
ALLOWED_ORIGINS=http://localhost:3000

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60

# Request Size
MAX_REQUEST_SIZE=10485760

# Prometheus Auth
PROMETHEUS_AUTH_USER=admin
PROMETHEUS_AUTH_PASSWORD=changeme

# Database
DB_QUERY_TIMEOUT=30
DB_LOG_LEVEL=info
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
```

### Security Headers Implemented

All responses include:
- ✅ `X-Frame-Options: DENY`
- ✅ `X-Content-Type-Options: nosniff`
- ✅ `X-XSS-Protection: 1; mode=block`
- ✅ `Strict-Transport-Security: max-age=31536000; includeSubDomains` (production)
- ✅ `Content-Security-Policy: default-src 'self'; ...`
- ✅ `Referrer-Policy: strict-origin-when-cross-origin`
- ✅ `Permissions-Policy: geolocation=(), microphone=(), camera=()`

### References
- [SECURITY.md](../SECURITY.md) - Complete security documentation
- [AUTHENTIK_SETUP.md](../AUTHENTIK_SETUP.md) - Authentication setup guide
- [.env.example](../.env.example) - Security configuration template
