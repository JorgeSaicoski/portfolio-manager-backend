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

