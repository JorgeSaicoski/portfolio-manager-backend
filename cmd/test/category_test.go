package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCategory_GetOwn tests getting user's own categories
func TestCategory_GetOwn(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_WithPagination", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		CreateTestCategory(testDB.DB, portfolio.ID, userID)
		CreateTestCategory(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", "/api/categories/own?page=1&limit=10", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			assert.Contains(t, body, "page")
			assert.Contains(t, body, "limit")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_EmptyList", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		resp := MakeRequest(t, "GET", "/api/categories/own?page=1&limit=10", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.Equal(t, 0, len(data))
		})
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/categories/own", nil, "")
		assert.Equal(t, 401, resp.Code)
	})

	t.Run("Pagination_LargePageNumber", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		CreateTestCategory(testDB.DB, portfolio.ID, userID) // Create 1 item

		resp := MakeRequest(t, "GET", "/api/categories/own?page=100&limit=10", nil, token)
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

		resp := MakeRequest(t, "GET", "/api/categories/own?page=1&limit=101", nil, token)
		assert.Equal(t, 200, resp.Code) // API doesn't validate max limit

		cleanDatabase(testDB.DB)
	})

	t.Run("Pagination_InvalidPageZero", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/categories/own?page=0&limit=10", nil, token)
		assert.Equal(t, 200, resp.Code) // API doesn't validate page=0
	})

	t.Run("Pagination_InvalidNegativePage", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/categories/own?page=-1&limit=10", nil, token)
		assert.Equal(t, 200, resp.Code) // API doesn't validate negative pages
	})

	t.Run("Pagination_BoundaryExactlyAtLimit", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		// Create exactly 10 items
		for i := 0; i < 10; i++ {
			CreateTestCategoryWithTitle(testDB.DB, portfolio.ID, userID, fmt.Sprintf("Category %d", i))
		}

		resp := MakeRequest(t, "GET", "/api/categories/own?page=1&limit=10", nil, token)
		AssertJSONResponse(t, resp, 200, func(data map[string]interface{}) {
			items := data["data"].([]interface{})
			assert.Equal(t, 10, len(items))
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Pagination_SecondPageWithRemainder", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		// Create 15 items (page 1: 10, page 2: 5)
		for i := 0; i < 15; i++ {
			CreateTestCategoryWithTitle(testDB.DB, portfolio.ID, userID, fmt.Sprintf("Category %d", i))
		}

		resp := MakeRequest(t, "GET", "/api/categories/own?page=2&limit=10", nil, token)
		AssertJSONResponse(t, resp, 200, func(data map[string]interface{}) {
			items := data["data"].([]interface{})
			assert.Equal(t, 5, len(items))
			assert.Equal(t, float64(2), data["page"])
		})

		cleanDatabase(testDB.DB)
	})
}

// TestCategory_Create tests creating a new category
func TestCategory_Create(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_WithAllFields", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "Web Development",
			"description":  "Web development projects",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/categories/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Web Development", data["title"])
			assert.Equal(t, "Web development projects", data["description"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_OnlyRequiredFields", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "Mobile Apps",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/categories/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Mobile Apps", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_EmptyTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/categories/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_MissingPortfolioID", func(t *testing.T) {
		payload := map[string]interface{}{
			"title": "Test Category",
		}

		resp := MakeRequest(t, "POST", "/api/categories/own", payload, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("ValidationError_InvalidPortfolioID", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":        "Test Category",
			"portfolio_id": 99999,
		}

		resp := MakeRequest(t, "POST", "/api/categories/own", payload, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "Test Category",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/categories/own", payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestCategory_GetByID tests getting a specific category
func TestCategory_GetByID(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_OwnCategory", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/categories/own/%d", category.ID), nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Test Category", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/categories/own/99999", nil, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/categories/own/%d", category.ID), nil, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestCategory_Update tests updating a category
func TestCategory_Update(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_UpdateTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "Updated Category",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/categories/own/%d", category.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Updated Category", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateDescription", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "Test Category",
			"description":  "Updated description",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/categories/own/%d", category.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Updated description", data["description"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_EmptyTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/categories/own/%d", category.ID), payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "Updated Category",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", "/api/categories/own/99999", payload, token)
		assert.Equal(t, 404, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "Updated Category",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/categories/own/%d", category.ID), payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestCategory_Delete tests deleting a category
func TestCategory_Delete(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_DeleteCategory", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/categories/own/%d", category.ID), nil, token)
		assert.Equal(t, 200, resp.Code)

		// Verify deletion
		getResp := MakeRequest(t, "GET", fmt.Sprintf("/api/categories/own/%d", category.ID), nil, token)
		assert.Equal(t, 404, getResp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "DELETE", "/api/categories/own/99999", nil, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/categories/own/%d", category.ID), nil, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_DeleteWithProjects_CascadeDelete", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/categories/own/%d", category.ID), nil, token)
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestCategory_GetPublic tests getting public categories
func TestCategory_GetPublic(t *testing.T) {
	userID := GetTestUserID()

	t.Run("Success_PublicCategory", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/categories/public/%d", category.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Test Category", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/categories/public/99999", nil, "")
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Success_NoAuthRequired", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/categories/public/%d", category.ID), nil, "")
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestCategory_GetByPortfolio tests getting categories by portfolio
func TestCategory_GetByPortfolio(t *testing.T) {
	userID := GetTestUserID()

	t.Run("Success_GetCategoriesByPortfolio", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		CreateTestCategory(testDB.DB, portfolio.ID, userID)
		CreateTestCategory(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/public/%d/categories", portfolio.ID), nil, "")

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

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/public/%d/categories", portfolio.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.Equal(t, 0, len(data))
		})

		cleanDatabase(testDB.DB)
	})
}
