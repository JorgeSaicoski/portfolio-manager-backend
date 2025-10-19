# Portfolio Manager API Documentation

## Base URL
```
http://localhost:8000/api
```

## Authentication
Most endpoints require JWT authentication. Include the token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Response Formats

### Success Response (with data)
```json
{
  "message": "Success message",
  "data": { ... }
}
```

### Success Response (paginated)
```json
{
  "data": [...],
  "page": 1,
  "limit": 10,
  "message": "Success"
}
```

### Error Response
```json
{
  "error": "Error message"
}
```

## Common Query Parameters

### Pagination
- `page` (integer, optional): Page number (default: 1, min: 1)
- `limit` (integer, optional): Items per page (default: 10, min: 1, max: 100)

---

## Portfolio Endpoints

### Get Own Portfolios
Returns a paginated list of portfolios owned by the authenticated user.

**Endpoint:** `GET /portfolios/own`

**Authentication:** Required

**Query Parameters:**
- `page` (optional): Page number
- `limit` (optional): Items per page

**Success Response (200):**
```json
{
  "data": [
    {
      "id": 1,
      "title": "My Portfolio",
      "description": "Portfolio description",
      "owner_id": "user123",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ],
  "page": 1,
  "limit": 10,
  "message": "Success"
}
```

**Error Responses:**
- `401 Unauthorized`: Missing or invalid authentication token
- `500 Internal Server Error`: Failed to retrieve portfolios

---

### Create Portfolio
Creates a new portfolio for the authenticated user.

**Endpoint:** `POST /portfolios/own`

**Authentication:** Required

**Request Body:**
```json
{
  "title": "My New Portfolio",
  "description": "Optional description"
}
```

**Validation:**
- `title`: Required, 1-255 characters
- `description`: Optional, max 1000 characters

**Success Response (201):**
```json
{
  "message": "Portfolio created successfully",
  "data": {
    "id": 2,
    "title": "My New Portfolio",
    "description": "Optional description",
    "owner_id": "user123",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Missing or invalid authentication token
- `500 Internal Server Error`: Failed to create portfolio

---

### Update Portfolio
Updates an existing portfolio owned by the authenticated user.

**Endpoint:** `PUT /portfolios/own/id/:id`

**Authentication:** Required

**Path Parameters:**
- `id` (integer): Portfolio ID

**Request Body:**
```json
{
  "title": "Updated Title",
  "description": "Updated description"
}
```

**Validation:**
- `title`: Optional, 1-255 characters
- `description`: Optional, max 1000 characters

**Success Response (200):**
```json
{
  "message": "Portfolio updated successfully",
  "data": {
    "id": 1,
    "title": "Updated Title",
    "description": "Updated description",
    "owner_id": "user123",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T12:00:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid portfolio ID or request data
- `401 Unauthorized`: Missing or invalid authentication token
- `500 Internal Server Error`: Failed to update portfolio

---

### Delete Portfolio
Deletes a portfolio owned by the authenticated user.

**Endpoint:** `DELETE /portfolios/own/id/:id`

**Authentication:** Required

**Path Parameters:**
- `id` (integer): Portfolio ID

**Success Response (200):**
```json
{
  "message": "Portfolio deleted successfully"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid portfolio ID
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: Not the owner of the portfolio
- `404 Not Found`: Portfolio not found
- `500 Internal Server Error`: Failed to delete portfolio

---

### Get Portfolio by ID (Public)
Retrieves a portfolio with all relationships (sections, categories, projects).

**Endpoint:** `GET /portfolios/id/:id`

**Authentication:** Not required (public endpoint)

**Path Parameters:**
- `id` (integer): Portfolio ID

**Success Response (200):**
```json
{
  "message": "Success",
  "data": {
    "id": 1,
    "title": "My Portfolio",
    "description": "Portfolio description",
    "owner_id": "user123",
    "sections": [
      {
        "id": 1,
        "title": "About",
        "description": "About section",
        "type": "text",
        "portfolio_id": 1,
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z"
      }
    ],
    "categories": [
      {
        "id": 1,
        "title": "Web Development",
        "description": "Web dev projects",
        "portfolio_id": 1,
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z"
      }
    ],
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid portfolio ID
- `404 Not Found`: Portfolio not found

---

## Category Endpoints

### Get Own Categories
Returns a paginated list of categories owned by the authenticated user.

**Endpoint:** `GET /categories/own`

**Authentication:** Required

**Query Parameters:**
- `page` (optional): Page number
- `limit` (optional): Items per page

**Success Response (200):**
```json
{
  "categories": [
    {
      "id": 1,
      "title": "Web Development",
      "description": "Web dev projects",
      "owner_id": "user123",
      "portfolio_id": 1,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ],
  "page": 1,
  "limit": 10,
  "message": "Success"
}
```

---

### Create Category
Creates a new category for the authenticated user.

**Endpoint:** `POST /categories/own`

**Authentication:** Required

**Request Body:**
```json
{
  "title": "Mobile Development",
  "description": "Mobile app projects",
  "portfolio_id": 1
}
```

**Validation:**
- `title`: Required, 1-255 characters
- `description`: Optional, max 1000 characters
- `portfolio_id`: Required, min 1

**Success Response (201):**
```json
{
  "category": {
    "id": 2,
    "title": "Mobile Development",
    "description": "Mobile app projects",
    "owner_id": "user123",
    "portfolio_id": 1,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "message": "Category created successfully"
}
```

---

### Update Category
Updates an existing category owned by the authenticated user.

**Endpoint:** `PUT /categories/own/id/:id`

**Authentication:** Required

**Request Body:**
```json
{
  "title": "Updated Category",
  "description": "Updated description",
  "portfolio_id": 1
}
```

---

### Delete Category
Deletes a category owned by the authenticated user.

**Endpoint:** `DELETE /categories/own/id/:id`

**Authentication:** Required

**Path Parameters:**
- `id` (integer): Category ID

---

### Get Category by ID (Public)
Retrieves a category with all projects.

**Endpoint:** `GET /categories/id/:id`

**Authentication:** Not required (public endpoint)

**Success Response (200):**
```json
{
  "category": {
    "id": 1,
    "title": "Web Development",
    "description": "Web dev projects",
    "owner_id": "user123",
    "portfolio_id": 1,
    "projects": [
      {
        "id": 1,
        "title": "E-commerce Platform",
        "main_image": "https://example.com/image.jpg",
        "description": "Full-stack e-commerce platform",
        "skills": ["React", "Node.js", "PostgreSQL"],
        "client": "ABC Company",
        "link": "https://example.com",
        "category_id": 1,
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z"
      }
    ],
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "message": "Success"
}
```

---

## Project Endpoints

### Create Project
Creates a new project for the authenticated user.

**Endpoint:** `POST /projects/`

**Authentication:** Required

**Request Body:**
```json
{
  "title": "E-commerce Platform",
  "images": ["https://example.com/img1.jpg", "https://example.com/img2.jpg"],
  "main_image": "https://example.com/main.jpg",
  "description": "Full-stack e-commerce platform built with modern technologies",
  "skills": ["React", "Node.js", "PostgreSQL", "Docker"],
  "client": "ABC Company",
  "link": "https://example.com",
  "category_id": 1
}
```

**Validation:**
- `title`: Required, 1-255 characters
- `images`: Optional array of URLs
- `main_image`: Optional, must be valid URL
- `description`: Required, min 1 character
- `skills`: Optional array of strings
- `client`: Optional, max 255 characters
- `link`: Optional, must be valid URL
- `category_id`: Required, min 1

**Success Response (201):**
```json
{
  "project": {
    "id": 1,
    "title": "E-commerce Platform",
    "images": ["https://example.com/img1.jpg"],
    "main_image": "https://example.com/main.jpg",
    "description": "Full-stack e-commerce platform",
    "skills": ["React", "Node.js"],
    "client": "ABC Company",
    "link": "https://example.com",
    "owner_id": "user123",
    "category_id": 1,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "message": "Project created successfully"
}
```

---

### Get Project by ID (Public)
Retrieves a single project.

**Endpoint:** `GET /projects/id/:id`

**Authentication:** Not required (public endpoint)

**Path Parameters:**
- `id` (integer): Project ID

---

### Get Projects by Category (Public)
Retrieves all projects in a specific category.

**Endpoint:** `GET /projects/category/:categoryId`

**Authentication:** Not required (public endpoint)

**Success Response (200):**
```json
{
  "projects": [...],
  "message": "Success"
}
```

---

### Search Projects by Skills (Public)
Searches for projects that match given skills.

**Endpoint:** `GET /projects/search/skills`

**Authentication:** Not required (public endpoint)

**Query Parameters:**
- `skills` (array): Array of skill names (e.g., `?skills=React&skills=Node.js`)

**Success Response (200):**
```json
{
  "projects": [...],
  "message": "Success"
}
```

**Error Responses:**
- `400 Bad Request`: At least one skill is required

---

### Search Projects by Client (Public)
Searches for projects by client name.

**Endpoint:** `GET /projects/search/client`

**Authentication:** Not required (public endpoint)

**Query Parameters:**
- `client` (string, required): Client name

**Success Response (200):**
```json
{
  "projects": [...],
  "message": "Success"
}
```

---

### Update Project
Updates an existing project owned by the authenticated user.

**Endpoint:** `PUT /projects/id/:id`

**Authentication:** Required

---

### Delete Project
Deletes a project owned by the authenticated user.

**Endpoint:** `DELETE /projects/id/:id`

**Authentication:** Required

---

## Section Endpoints

### Create Section
Creates a new section for the authenticated user.

**Endpoint:** `POST /sections/`

**Authentication:** Required

**Request Body:**
```json
{
  "title": "About Me",
  "description": "Information about myself",
  "type": "text",
  "portfolio_id": 1
}
```

**Validation:**
- `title`: Required, 1-255 characters
- `description`: Optional, max 1000 characters
- `type`: Required, 1-100 characters
- `portfolio_id`: Required, min 1

---

### Get Section by ID (Public)
Retrieves a section with relationships.

**Endpoint:** `GET /sections/id/:id`

**Authentication:** Not required (public endpoint)

---

### Get Sections by Portfolio (Public)
Retrieves all sections for a portfolio.

**Endpoint:** `GET /sections/portfolio/:portfolioId`

**Authentication:** Not required (public endpoint)

---

### Get Sections by Type (Public)
Retrieves all sections of a specific type.

**Endpoint:** `GET /sections/type`

**Authentication:** Not required (public endpoint)

**Query Parameters:**
- `type` (string, required): Section type

**Error Responses:**
- `400 Bad Request`: Section type is required

---

### Update Section
Updates an existing section owned by the authenticated user.

**Endpoint:** `PUT /sections/id/:id`

**Authentication:** Required

---

### Delete Section
Deletes a section owned by the authenticated user.

**Endpoint:** `DELETE /sections/id/:id`

**Authentication:** Required

---

## HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters or body
- `401 Unauthorized`: Authentication required or invalid token
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

---

## Data Transfer Objects (DTOs)

The API uses DTOs for request and response validation. All DTOs are located in:
- Request DTOs: `backend/internal/dto/request/`
- Response DTOs: `backend/internal/dto/response/`

### Request DTO Validation
Request DTOs include validation tags using Gin's binding framework:
- `required`: Field must be present
- `min`: Minimum value/length
- `max`: Maximum value/length
- `url`: Must be valid URL
- `email`: Must be valid email

### Response DTO Structure
Response DTOs provide consistent structure and exclude sensitive database fields like soft-deleted timestamps.
