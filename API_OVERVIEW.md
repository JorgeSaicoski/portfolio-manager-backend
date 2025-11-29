# Portfolio Manager API - Complete Reference

> **For AI Assistants & Frontend Developers**
> This is the definitive guide to all backend API endpoints. No need to investigate backend code.

**Quick Links:**
- [Authentication](#authentication) | [Quick Start](#quick-start) | [Response Formats](#response-formats)
- [Portfolios](#portfolios) | [Categories](#categories) | [Projects](#projects) | [Sections](#sections)
- [Section Contents](#section-contents) | [Images](#images) | [Users](#users)

**Related Documentation:**
- [Image API Details](/docs/api/images.md) - Comprehensive image management guide
- [Authentication Setup](/docs/authentication/) - OAuth2/OIDC configuration

---

## Overview

**Base URL:** `http://localhost:8000/api`

**Technology:**
- **Backend:** Go 1.24.0 with Gin framework
- **Database:** PostgreSQL with GORM ORM
- **Authentication:** OAuth2/OIDC via Authentik (Bearer JWT tokens)
- **Features:** Pagination, image optimization, audit logging, Prometheus metrics

### Access Levels

**🔒 Admin (Authenticated):** Manage your own portfolios
- Endpoints: `/api/{resource}/own/*`
- Required: `Authorization: Bearer <JWT_TOKEN>`
- User identified via JWT `sub` claim (userID)
- Users can only access/modify their own data

**🌐 Public (Visitors):** View published portfolios
- Endpoints: `/api/{resource}/public/:id` or `/api/{resource}/id/:id`
- No authentication required
- Read-only access

### Quick Start

1. **Get JWT Token** (via Authentik login or token endpoint)
2. **Make Request:**
   ```bash
   curl -H "Authorization: Bearer YOUR_TOKEN" \
        http://localhost:8000/api/portfolios/own
   ```
3. **See responses in standardized format** (details below)

---

## Response Formats

### Success (200/201)
```json
{
  "data": { /* resource or array */ },
  "message": "Success message"
}
```

### Success with Pagination (200)
```json
{
  "data": [ /* array of items */ ],
  "page": 1,
  "limit": 10,
  "message": "Success"
}
```

### Error (4xx/5xx)
```json
{
  "error": "Human-readable error message"
}
```

---

## Pagination

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1, min: 1)
- `limit` (integer, optional): Items per page (default: 10, min: 1, max: 100)

**Example:**
```bash
GET /api/portfolios/own?page=2&limit=20
```

**Applied to:** All list endpoints (`GET /own`, `GET /public/:id/categories`, etc.)

---

## Authentication

**Method:** Bearer JWT tokens from Authentik OAuth2/OIDC provider

**Header:**
```
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Token Contains:**
- `sub`: User ID (used as `userID` in backend)
- `email`: User email
- `preferred_username`: Username
- `name`, `given_name`, `family_name`: User profile info

**Ownership Model:**
- All resources have `owner_id` field
- Users can only modify/delete their own resources
- Attempting to modify others' resources → `403 Forbidden`

**See Also:** [Authentication Setup Docs](/docs/authentication/)

---

## Common Patterns

### Position & Ordering
- Categories and Sections have `position` field for custom ordering
- Update single position: `PUT /categories/own/:id/position`
- Bulk reorder: `PUT /categories/own/reorder` (array of {id, position})

### Image Handling
- Images use polymorphic association (`entity_type`, `entity_id`)
- Automatic optimization: max 1920px width, 85% JPEG quality
- Thumbnail generation: 400px width
- See [Image API Guide](/docs/api/images.md)

### Soft Deletes
- Resources support soft deletion (GORM DeletedAt)
- Deleted resources excluded from queries
- Hard delete after retention period (configurable)

### Error Codes
- `400 Bad Request`: Invalid input/validation failure
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: Valid auth but access denied (not owner)
- `404 Not Found`: Resource doesn't exist
- `500 Internal Server Error`: Server-side error (logged)

---

## Portfolios

Portfolios are the top-level container for a user's work. Each portfolio contains categories, sections, and projects.

### Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/portfolios/own` | 🔒 | List authenticated user's portfolios (paginated) |
| POST | `/api/portfolios/own` | 🔒 | Create new portfolio |
| GET | `/api/portfolios/own/:id` | 🔒 | Get own portfolio by ID |
| PUT | `/api/portfolios/own/:id` | 🔒 | Update portfolio (title, description) |
| DELETE | `/api/portfolios/own/:id` | 🔒 | Delete portfolio (cascades to all related data) |
| GET | `/api/portfolios/id/:id` | 🌐 | Get portfolio by ID (public view with nested data) |
| GET | `/api/portfolios/public/:id` | 🌐 | Get portfolio by ID (alias for `/id/:id`) |
| GET | `/api/portfolios/public/:id/categories` | 🌐 | Get all categories in portfolio |
| GET | `/api/portfolios/public/:id/sections` | 🌐 | Get all sections in portfolio |

### Request/Response Details

**Create Portfolio (POST /own):**
```json
// Request
{
  "title": "My Portfolio",
  "description": "Optional description"
}

// Response (201)
{
  "data": {
    "id": 1,
    "title": "My Portfolio",
    "description": "Optional description",
    "owner_id": "user-123",
    "created_at": "2025-11-29T10:00:00Z",
    "updated_at": "2025-11-29T10:00:00Z"
  },
  "message": "Portfolio created successfully"
}
```

**Get Public Portfolio (GET /public/:id):**
- Returns portfolio with nested `sections[]` and `categories[]` arrays
- Useful for rendering full portfolio view

**Notes:**
- Deleting a portfolio cascades to all categories, sections, projects, and section contents
- Each user can have multiple portfolios

---

## Categories

Categories organize projects within a portfolio (e.g., "Web Development", "Mobile Apps").

### Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/categories/own` | 🔒 | List authenticated user's categories (paginated) |
| POST | `/api/categories/own` | 🔒 | Create new category |
| GET | `/api/categories/own/:id` | 🔒 | Get own category by ID |
| PUT | `/api/categories/own/:id` | 🔒 | Update category (title, description, portfolio_id) |
| PUT | `/api/categories/own/:id/position` | 🔒 | Update single category position |
| PUT | `/api/categories/own/reorder` | 🔒 | Bulk reorder categories |
| DELETE | `/api/categories/own/:id` | 🔒 | Delete category (cascades to projects) |
| GET | `/api/categories/id/:id` | 🌐 | Get category by ID (public view) |
| GET | `/api/categories/public/:id` | 🌐 | Get category by ID (alias) |
| GET | `/api/categories/public/:id/projects` | 🌐 | Get all projects in category |

### Request/Response Details

**Create Category (POST /own):**
```json
// Request
{
  "title": "Web Development",
  "description": "Full-stack web projects",
  "portfolio_id": 1
}

// Validation
// - title: required, 1-255 chars
// - description: optional, max 1000 chars
// - portfolio_id: required, must be owned by user
```

**Update Position (PUT /own/:id/position):**
```json
// Request
{
  "position": 3
}
```

**Bulk Reorder (PUT /own/reorder):**
```json
// Request
{
  "categories": [
    {"id": 1, "position": 0},
    {"id": 3, "position": 1},
    {"id": 2, "position": 2}
  ]
}
```

**Notes:**
- Categories have custom ordering via `position` field
- Public endpoints return categories with nested projects

---

## Projects

Projects represent individual work items within categories.

### Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/projects/own` | 🔒 | List authenticated user's projects (paginated) |
| POST | `/api/projects/own` | 🔒 | Create new project |
| GET | `/api/projects/own/:id` | 🔒 | Get own project by ID |
| PUT | `/api/projects/own/:id` | 🔒 | Update project |
| DELETE | `/api/projects/own/:id` | 🔒 | Delete project |
| GET | `/api/projects/public/:id` | 🌐 | Get project by ID (public view) |
| GET | `/api/projects/category/:categoryId` | 🌐 | Get all projects in category |
| GET | `/api/projects/search/skills` | 🌐 | Search projects by skills |
| GET | `/api/projects/search/client` | 🌐 | Search projects by client name |

### Request/Response Details

**Create Project (POST /own):**
```json
// Request
{
  "title": "E-commerce Platform",
  "description": "Full-stack e-commerce site",
  "main_image": "https://example.com/image.jpg",
  "images": ["https://example.com/img1.jpg", "https://example.com/img2.jpg"],
  "skills": ["React", "Node.js", "PostgreSQL"],
  "client": "ABC Company",
  "link": "https://example.com",
  "category_id": 1
}

// Validation
// - title: required, 1-255 chars
// - description: required, min 1 char
// - main_image: optional, must be valid URL
// - images: optional array of URLs
// - skills: optional array of strings
// - client: optional, max 255 chars
// - link: optional, must be valid URL
// - category_id: required, must be owned by user
```

**Search by Skills (GET /search/skills):**
```bash
GET /api/projects/search/skills?skills=React&skills=Node.js
# Returns projects matching ANY of the skills
```

**Search by Client (GET /search/client):**
```bash
GET /api/projects/search/client?client=ABC%20Company
```

**Notes:**
- Skills stored as JSON array in database
- Main image can be set for gallery/list views

---

## Sections

Sections are custom content areas in a portfolio (e.g., "About Me", "Skills", "Contact").

### Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/sections/own` | 🔒 | List authenticated user's sections (paginated) |
| POST | `/api/sections/own` | 🔒 | Create new section |
| GET | `/api/sections/own/:id` | 🔒 | Get own section by ID |
| PUT | `/api/sections/own/:id` | 🔒 | Update section |
| PUT | `/api/sections/own/:id/position` | 🔒 | Update single section position |
| PUT | `/api/sections/own/reorder` | 🔒 | Bulk reorder sections |
| DELETE | `/api/sections/own/:id` | 🔒 | Delete section (cascades to section contents) |
| GET | `/api/sections/public/:id` | 🌐 | Get section by ID (public view) |
| GET | `/api/sections/portfolio/:portfolioId` | 🌐 | Get all sections for portfolio |
| GET | `/api/sections/type` | 🌐 | Get sections by type (query param) |

### Request/Response Details

**Create Section (POST /own):**
```json
// Request
{
  "title": "About Me",
  "description": "Personal introduction",
  "type": "text",
  "portfolio_id": 1
}

// Validation
// - title: required, 1-255 chars
// - description: optional, max 1000 chars
// - type: required, 1-100 chars (e.g., "text", "gallery", "timeline")
// - portfolio_id: required, must be owned by user
```

**Get by Type (GET /type):**
```bash
GET /api/sections/type?type=gallery
# Returns all sections with type="gallery"
```

**Notes:**
- Sections have custom ordering via `position` field
- Type field allows flexible section categorization
- Section contents are separate resources (see below)

---

## Section Contents

Section contents are individual content blocks within a section (text, images, code blocks, etc.).

### Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/section-contents/own` | 🔒 | Create new section content block |
| PUT | `/api/section-contents/own/:id` | 🔒 | Update section content |
| PATCH | `/api/section-contents/own/:id/order` | 🔒 | Update content block order |
| DELETE | `/api/section-contents/own/:id` | 🔒 | Delete section content |
| GET | `/api/section-contents/:id` | 🌐 | Get section content by ID |
| GET | `/api/sections/:sectionId/contents` | 🌐 | Get all contents for section |

### Request/Response Details

**Create Section Content (POST /own):**
```json
// Request
{
  "section_id": 1,
  "type": "text",
  "content": "This is my about section...",
  "order": 0,
  "image_id": null
}

// Validation
// - section_id: required, section must be in user's portfolio
// - type: required (e.g., "text", "image", "code", "quote")
// - content: optional, depends on type
// - order: optional, for ordering blocks within section
// - image_id: optional, references uploaded image
```

**Update Order (PATCH /own/:id/order):**
```json
// Request
{
  "order": 2
}
```

**Get Section Contents (GET /sections/:sectionId/contents):**
- Returns array of content blocks ordered by `order` field
- Public endpoint, no auth required

**Notes:**
- Ownership validated via section → portfolio → owner_id chain
- Can reference images via `image_id` for image galleries
- Deleting associated image cleans up image_id reference

---

## Images

Images can be attached to portfolios, projects, or sections. See [Image API Guide](/docs/api/images.md) for full details.

### Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/images/own` | 🔒 | Upload image (multipart/form-data) |
| GET | `/api/images/own` | 🔒 | Get authenticated user's images |
| PUT | `/api/images/own/:id` | 🔒 | Update image metadata (alt, is_main) |
| DELETE | `/api/images/own/:id` | 🔒 | Delete image and files |
| GET | `/api/images/entity/:type/:id` | 🌐 | Get images for entity (type: project/portfolio/section) |

### Request/Response Details

**Upload Image (POST /own):**
```bash
# multipart/form-data
curl -X POST http://localhost:8000/api/images/own \
  -H "Authorization: Bearer TOKEN" \
  -F "file=@image.jpg" \
  -F "entity_type=project" \
  -F "entity_id=1" \
  -F "alt=Project screenshot" \
  -F "is_main=true"

# Validation
# - file: required, JPEG/PNG/WebP, max 10MB
# - entity_type: required, one of: project, portfolio, section
# - entity_id: required, entity must be owned by user
# - alt: optional, accessibility text
# - is_main: optional, boolean (default: false)
```

**Response (201):**
```json
{
  "data": {
    "id": 1,
    "url": "/uploads/images/original/abc123_image.jpg",
    "thumbnail_url": "/uploads/images/thumbnail/abc123_image.jpg",
    "file_name": "image.jpg",
    "file_size": 245678,
    "mime_type": "image/jpeg",
    "alt": "Project screenshot",
    "entity_type": "project",
    "entity_id": 1,
    "is_main": true,
    "owner_id": "user-123",
    "created_at": "2025-11-29T10:00:00Z",
    "updated_at": "2025-11-29T10:00:00Z"
  },
  "message": "Image uploaded successfully"
}
```

**Features:**
- Automatic resize to 1920px max width (85% JPEG quality)
- Automatic thumbnail generation (400px width)
- File validation (type, size)
- Ownership validation (entity must belong to user)
- Audit logging

**Static Files:**
- Uploaded images served at `/uploads/images/original/...`
- Thumbnails served at `/uploads/images/thumbnail/...`

**See Also:** [Complete Image API Documentation](/docs/api/images.md)

---

## Users

User management endpoints for GDPR compliance and data summary.

### Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/users/me/summary` | 🔒 | Get summary of user's data |
| DELETE | `/api/users/me/data` | 🔒 | Delete all user data (GDPR compliance) |

### Request/Response Details

**Get Data Summary (GET /me/summary):**
```json
// Response (200)
{
  "data": {
    "userID": "user-123",
    "portfolios": 2,
    "categories": 5,
    "sections": 8,
    "projects": 12,
    "totalItems": 27
  },
  "message": "User data summary"
}
```

**Delete All Data (DELETE /me/data):**
- Deletes all portfolios owned by user
- CASCADE deletes all categories, sections, projects, section_contents
- Returns count of deleted portfolios

```json
// Response (200)
{
  "message": "User data cleaned up successfully",
  "portfoliosDeleted": 2
}
```

**Notes:**
- Data summary useful for showing users what will be deleted
- Delete operation is permanent (no soft delete for user cleanup)
- Intended for GDPR "right to be forgotten" compliance

---

## Additional Endpoints

### Health & Monitoring

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | None | Health check (status + DB connection) |
| GET | `/ready` | None | Readiness probe for K8s |
| HEAD | `/health` | None | Quick health check (no body) |
| GET | `/metrics` | Basic Auth | Prometheus metrics (optional auth) |

**Health Check Response:**
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2025-11-29T10:00:00Z"
}
```

**Metrics:**
- Protected with Basic Auth if `PROMETHEUS_AUTH_USER` and `PROMETHEUS_AUTH_PASSWORD` set
- Exposes Gin metrics, DB connection pool stats, custom business metrics
- Format: Prometheus text-based exposition format

### Static Files

| Path | Description |
|------|-------------|
| `/uploads/images/original/*` | Full-size optimized images |
| `/uploads/images/thumbnail/*` | 400px thumbnails |

**Notes:**
- Static files served directly by Gin
- No authentication required (public access)
- Files stored in Docker volume for persistence

---

## Error Handling

### HTTP Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful GET/PUT/DELETE |
| 201 | Created | Successful POST |
| 400 | Bad Request | Invalid input, validation failure, missing required fields |
| 401 | Unauthorized | Missing/invalid token, token expired |
| 403 | Forbidden | Valid auth but access denied (not owner) |
| 404 | Not Found | Resource doesn't exist |
| 500 | Internal Server Error | Database error, file system error, unexpected error |

### Error Response Format

All errors return JSON with single `error` field:
```json
{
  "error": "Human-readable error message"
}
```

### Common Error Scenarios

**401 Unauthorized:**
```json
{
  "error": "Authorization header required"
}
{
  "error": "Invalid token"
}
```

**403 Forbidden (Ownership):**
```json
{
  "error": "Access denied: portfolio belongs to another user"
}
```

**400 Bad Request (Validation):**
```json
{
  "error": "Invalid request data"
}
{
  "error": "Invalid portfolio ID"
}
{
  "error": "Section type is required"
}
```

**404 Not Found:**
```json
{
  "error": "Portfolio not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Failed to retrieve portfolios"
}
{
  "error": "Failed to save image"
}
```

### Error Logging

- All 4xx/5xx responses logged to audit system
- 5xx errors include detailed stack traces (not returned to client)
- Logs stored in `/backend/audit/` directory
- Separate log files: `create.log`, `update.log`, `delete.log`, `error.log`

---

## Best Practices

### Authentication
- Always include `Authorization: Bearer <token>` header for protected endpoints
- Tokens expire after configured period (check with Authentik admin)
- Refresh tokens before expiry to avoid interruptions
- Never share tokens in URLs or logs

### Pagination
- Use pagination for list endpoints to avoid large responses
- Start with reasonable `limit` (10-50) and adjust based on UI needs
- Cache results when appropriate to reduce API calls

### Image Upload
- Validate images client-side before upload (type, size)
- Use multipart/form-data encoding
- Set `is_main=true` for primary images
- Provide `alt` text for accessibility
- See [Image API Guide](/docs/api/images.md) for optimization details

### Error Handling
- Always check HTTP status code
- Parse `error` field for user-friendly messages
- Log detailed errors for debugging
- Implement retry logic for 5xx errors (with backoff)

### Performance
- Use public endpoints when authentication not needed
- Leverage HTTP caching headers (sent by backend)
- Request only needed data (avoid deep nesting when possible)
- Consider GraphQL for complex queries (future enhancement)

### Security
- Validate all user input client-side
- Never trust client-side validation alone (backend validates)
- Sanitize HTML/markdown content before rendering
- Use HTTPS in production
- Implement CSRF protection for browser-based apps

---

## Appendix

### Complete Endpoint Count

| Resource | Admin Endpoints | Public Endpoints | Total |
|----------|-----------------|------------------|-------|
| Portfolios | 5 | 4 | 9 |
| Categories | 7 | 3 | 10 |
| Projects | 5 | 4 | 9 |
| Sections | 7 | 3 | 10 |
| Section Contents | 4 | 2 | 6 |
| Images | 4 | 1 | 5 |
| Users | 2 | 0 | 2 |
| Health/Monitoring | 0 | 4 | 4 |
| **TOTAL** | **34** | **21** | **55** |

### Environment Variables

Key variables affecting API behavior:

| Variable | Purpose | Default |
|----------|---------|---------|
| `PORT` | Server port | 8000 |
| `AUTHENTIK_URL` | Authentik OIDC provider URL | Required |
| `AUTHENTIK_ISSUER` | Public issuer URL for tokens | Required |
| `TESTING_MODE` | Bypass auth for testing | false |
| `PROMETHEUS_AUTH_USER` | Metrics endpoint user | (optional) |
| `PROMETHEUS_AUTH_PASSWORD` | Metrics endpoint password | (optional) |
| `LOG_LEVEL` | Logging verbosity | info |

### Data Model Relationships

```
User (via Authentik)
  └── Portfolio (owner_id)
      ├── Category
      │   └── Project
      ├── Section
      │   └── Section Content
      │       └── Image (optional)
      └── Image (polymorphic)
```

**Cascade Behavior:**
- Delete Portfolio → deletes Categories, Sections, Projects, Section Contents
- Delete Category → deletes Projects
- Delete Section → deletes Section Contents
- Delete Image → nullifies image_id in Section Contents

### Related Documentation

- **Setup & Deployment:**
  - [Main README](/README.md)
  - [Setup Guide](/SETUP.md)
  - [Deployment Docs](/docs/deployment/)

- **API Details:**
  - [Image API Guide](/docs/api/images.md) - Comprehensive image management
  - [Existing Endpoint Docs](/docs/api/endpoints.md) - Partial endpoint reference

- **Authentication:**
  - [Authentication Setup](/docs/authentication/)
  - [Grafana SSO Setup](/docs/GRAFANA_SSO_SETUP.md)
  - [Troubleshooting](/docs/authentication/troubleshooting.md)

- **Operations:**
  - [Makefile Guide](/docs/MAKEFILE_GUIDE.md)
  - [Monitoring & Alerts](/docs/MONITORING_ALERTS.md)

### Support

- **Repository:** https://github.com/JorgeSaicoski/portfolio-manager
- **Issues:** https://github.com/JorgeSaicoski/portfolio-manager/issues
- **Documentation:** `/docs` directory
