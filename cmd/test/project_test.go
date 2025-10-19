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
			"images":      []string{"https://example.com/img1.png", "https://example.com/img2.png"},
			"main_image":  "https://example.com/main.png",
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

	t.Run("Success_WithImagesArray", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":       "Image Gallery Project",
			"description": "Project with multiple images",
			"images":      []string{"img1.png", "img2.png", "img3.png"},
			"category_id": category.ID,
		}

		resp := MakeRequest(t, "POST", "/api/projects/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			images := data["images"].([]interface{})
			assert.Equal(t, 3, len(images))
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
