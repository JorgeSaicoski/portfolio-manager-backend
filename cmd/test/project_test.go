package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProject_GetOwn tests getting user's own projects
func TestProject_GetOwn(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_WithPagination", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		CreateTestProject(testDB.DB, category.ID, userID)
		CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "GET", "/api/projects/own?page=1&limit=10", nil, token)

		// Use AssertPaginatedResponse helper for coverage
		AssertPaginatedResponse(t, resp)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			assert.Contains(t, body, "page")
			assert.Contains(t, body, "limit")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_EmptyList", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		resp := MakeRequest(t, "GET", "/api/projects/own?page=1&limit=10", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.Equal(t, 0, len(data))
		})
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/projects/own", nil, "")
		assert.Equal(t, 401, resp.Code)
	})

	t.Run("Pagination_LargePageNumber", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		CreateTestProject(testDB.DB, category.ID, userID) // Create 1 item

		resp := MakeRequest(t, "GET", "/api/projects/own?page=100&limit=10", nil, token)
		assert.Equal(t, 200, resp.Code)

		AssertJSONResponse(t, resp, 200, func(data map[string]interface{}) {
			items := data["data"].([]interface{})
			assert.Equal(t, 0, len(items)) // No items on page 100
			assert.Equal(t, float64(100), data["page"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Pagination_ExceedMaxLimit", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		resp := MakeRequest(t, "GET", "/api/projects/own?page=1&limit=101", nil, token)
		assert.Equal(t, 200, resp.Code) // API doesn't validate max limit

		cleanDatabase(testDB.DB)
	})

	t.Run("Pagination_InvalidPageZero", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/projects/own?page=0&limit=10", nil, token)
		assert.Equal(t, 200, resp.Code) // API doesn't validate page=0
	})

	t.Run("Pagination_InvalidNegativePage", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/projects/own?page=-1&limit=10", nil, token)
		assert.Equal(t, 200, resp.Code) // API doesn't validate negative pages
	})

	t.Run("Pagination_BoundaryExactlyAtLimit", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		// Create exactly 10 items
		for i := 0; i < 10; i++ {
			CreateTestProjectWithTitle(testDB.DB, category.ID, userID, fmt.Sprintf("Project %d", i))
		}

		resp := MakeRequest(t, "GET", "/api/projects/own?page=1&limit=10", nil, token)
		AssertJSONResponse(t, resp, 200, func(data map[string]interface{}) {
			items := data["data"].([]interface{})
			assert.Equal(t, 10, len(items))
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Pagination_SecondPageWithRemainder", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		// Create 15 items (page 1: 10, page 2: 5)
		for i := 0; i < 15; i++ {
			CreateTestProjectWithTitle(testDB.DB, category.ID, userID, fmt.Sprintf("Project %d", i))
		}

		resp := MakeRequest(t, "GET", "/api/projects/own?page=2&limit=10", nil, token)
		AssertJSONResponse(t, resp, 200, func(data map[string]interface{}) {
			items := data["data"].([]interface{})
			assert.Equal(t, 5, len(items))
			assert.Equal(t, float64(2), data["page"])
		})

		cleanDatabase(testDB.DB)
	})
}

// TestProject_Create tests creating a new project
func TestProject_Create(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_WithAllFields", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":       "E-commerce Platform",
			"description": "A full-stack e-commerce platform",
			"skills":      []string{"React", "Node.js", "PostgreSQL"},
			"client":      "Tech Corp",
			"link":        "https://example.com",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "E-commerce Platform", data["title"])
			assert.Equal(t, "A full-stack e-commerce platform", data["description"])
			assert.Equal(t, "Tech Corp", data["client"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_OnlyRequiredFields", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":       "Simple Project",
			"description": "A simple project",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Simple Project", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_EmptyTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":       "",
			"description": "Test description",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_MissingCategoryID", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Test Project",
			"description": "Test description",
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("ValidationError_InvalidCategoryID", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Test Project",
			"description": "Test description",
			"category_id": 99999,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Success_WithSkillsArray", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":       "Multi-skill Project",
			"description": "Project with multiple skills",
			"skills":      []string{"Go", "React", "Docker", "PostgreSQL"},
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			skills := data["skills"].([]interface{})
			assert.Equal(t, 4, len(skills))
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_WithoutImages", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":       "Project Without Images",
			"description": "A project with no images",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Project Without Images", data["title"])

			// Images should be an empty array or nil
			if images, ok := data["images"]; ok {
				if images != nil {
					imageArray := images.([]interface{})
					assert.Equal(t, 0, len(imageArray))
				}
			}
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":       "Test Project",
			"description": "Test description",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestProject_GetByID tests getting a specific project
func TestProject_GetByID(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_OwnProject", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/projects/own/%d", project.ID), nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Test Project", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/projects/own/99999", nil, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/projects/own/%d", project.ID), nil, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestProject_Update tests updating a project
func TestProject_Update(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_UpdateTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		payload := map[string]interface{}{
			"title":       "Updated Project",
			"description": "Test project description",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/projects/own/%d", project.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Updated Project", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateDescription", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		payload := map[string]interface{}{
			"title":       "Test Project",
			"description": "Updated description",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/projects/own/%d", project.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Updated description", data["description"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateSkills", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		payload := map[string]interface{}{
			"title":       "Test Project",
			"description": "Test project description",
			"skills":      []string{"Python", "Django", "Vue.js"},
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/projects/own/%d", project.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			skills := data["skills"].([]interface{})
			assert.Equal(t, 3, len(skills))
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateClient", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		payload := map[string]interface{}{
			"title":       "Test Project",
			"description": "Test project description",
			"client":      "New Client Inc",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/projects/own/%d", project.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "New Client Inc", data["client"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_EmptyTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		payload := map[string]interface{}{
			"title":       "",
			"description": "Test description",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/projects/own/%d", project.ID), payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":       "Updated Project",
			"description": "Test description",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "PUT", "/api/projects/own/99999", payload, token)
		assert.Equal(t, 404, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		payload := map[string]interface{}{
			"title":       "Updated Project",
			"description": "Test description",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/projects/own/%d", project.ID), payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestProject_Delete tests deleting a project
func TestProject_Delete(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_DeleteProject", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/projects/own/%d", project.ID), nil, token)
		assert.Equal(t, 200, resp.Code)

		// Verify deletion
		getResp := MakeRequest(t, "GET", fmt.Sprintf("/api/projects/own/%d", project.ID), nil, token)
		assert.Equal(t, 404, getResp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "DELETE", "/api/projects/own/99999", nil, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/projects/own/%d", project.ID), nil, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestProject_GetPublic tests getting public projects
func TestProject_GetPublic(t *testing.T) {
	userID := GetTestUserID()

	t.Run("Success_PublicProject", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/projects/public/%d", project.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Test Project", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/projects/public/99999", nil, "")
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Success_NoAuthRequired", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/projects/public/%d", project.ID), nil, "")
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestProject_GetByCategory tests getting projects by category
func TestProject_GetByCategory(t *testing.T) {
	userID := GetTestUserID()

	t.Run("Success_GetProjectsByCategory", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		CreateTestProject(testDB.DB, category.ID, userID)
		CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/categories/public/%d/projects", category.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.GreaterOrEqual(t, len(data), 2)
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_EmptyList", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/categories/public/%d/projects", category.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.Equal(t, 0, len(data))
		})

		cleanDatabase(testDB.DB)
	})
}

// TestProject_DuplicateCheck tests duplicate title validation within same category
func TestProject_DuplicateCheck(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Error_CreateDuplicate", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		// Create first project
		CreateTestProjectWithTitle(testDB.DB, category.ID, userID, "E-commerce Platform")

		// Try to create second project with same title in same category
		payload := map[string]interface{}{
			"title":       "E-commerce Platform",
			"description": "Another e-commerce project",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		AssertJSONResponse(t, resp, 400, func(body map[string]interface{}) {
			assert.Contains(t, body, "error")
			assert.Contains(t, body["error"], "already exists")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_CreateSameTitleDifferentCategory", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category1 := CreateTestCategoryWithTitle(testDB.DB, portfolio.ID, userID, "Web Development")
		category2 := CreateTestCategoryWithTitle(testDB.DB, portfolio.ID, userID, "Mobile Development")

		// Create project in first category
		CreateTestProjectWithTitle(testDB.DB, category1.ID, userID, "E-commerce Platform")

		// Create project with same title but in different category - should succeed
		payload := map[string]interface{}{
			"title":       "E-commerce Platform",
			"description": "Mobile version",
			"category_id": category2.ID,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)
		assert.Equal(t, 201, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Error_UpdateToDuplicate", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		// Create two projects
		project1 := CreateTestProjectWithTitle(testDB.DB, category.ID, userID, "Project One")
		project2 := CreateTestProjectWithTitle(testDB.DB, category.ID, userID, "Project Two")

		// Try to update project2 to have same title as project1
		payload := map[string]interface{}{
			"title":       project1.Title,
			"description": "Updated description",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/projects/own/%d", project2.ID), payload, token)
		assert.Equal(t, 400, resp.Code)

		AssertJSONResponse(t, resp, 400, func(body map[string]interface{}) {
			assert.Contains(t, body, "error")
			assert.Contains(t, body["error"], "already exists")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateSameTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		// Create project
		project := CreateTestProjectWithTitle(testDB.DB, category.ID, userID, "My Project")

		// Update with same title but different description - should succeed
		payload := map[string]interface{}{
			"title":       "My Project",
			"description": "Updated description",
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/projects/own/%d", project.ID), payload, token)
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestProject_Create_SetsPositionAutomatically tests that the database trigger
// automatically assigns sequential positions to projects within the same category
func TestProject_Create_SetsPositionAutomatically(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_AutoAssignSequentialPositions", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create test data
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		// Create first project without specifying position
		payload1 := map[string]interface{}{
			"title":       "First Project",
			"description": "First project in category",
			"category_id": category.ID,
		}
		resp1 := MakeRequest(t, "POST", "/api/projects/own", payload1, token)

		AssertJSONResponse(t, resp1, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "First Project", data["title"])
			// Position should be auto-assigned as 1
			position := data["position"].(float64)
			assert.Equal(t, float64(1), position, "First project should have position 1")
		})

		// Create second project without specifying position
		payload2 := map[string]interface{}{
			"title":       "Second Project",
			"description": "Second project in category",
			"category_id": category.ID,
		}
		resp2 := MakeRequest(t, "POST", "/api/projects/own", payload2, token)

		AssertJSONResponse(t, resp2, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Second Project", data["title"])
			// Position should be auto-assigned as 2
			position := data["position"].(float64)
			assert.Equal(t, float64(2), position, "Second project should have position 2")
		})

		// Create third project without specifying position
		payload3 := map[string]interface{}{
			"title":       "Third Project",
			"description": "Third project in category",
			"category_id": category.ID,
		}
		resp3 := MakeRequest(t, "POST", "/api/projects/own", payload3, token)

		AssertJSONResponse(t, resp3, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Third Project", data["title"])
			// Position should be auto-assigned as 3
			position := data["position"].(float64)
			assert.Equal(t, float64(3), position, "Third project should have position 3")
		})

		// Verify that projects in different categories get independent positions
		category2 := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		payload4 := map[string]interface{}{
			"title":       "First in Different Category",
			"description": "Should have position 1",
			"category_id": category2.ID,
		}
		resp4 := MakeRequest(t, "POST", "/api/projects/own", payload4, token)

		AssertJSONResponse(t, resp4, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			// Position should be 1 because it's a different category
			position := data["position"].(float64)
			assert.Equal(t, float64(1), position, "First project in different category should have position 1")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_ManualPositionRespected", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create test data
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		// Create project with explicit position
		payload := map[string]interface{}{
			"title":       "Project with Manual Position",
			"description": "Explicit position 5",
			"category_id": category.ID,
			"position":    5,
		}
		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			// Manual position should be respected
			position := data["position"].(float64)
			assert.Equal(t, float64(5), position, "Manual position should be respected")
		})

		cleanDatabase(testDB.DB)
	})
}
