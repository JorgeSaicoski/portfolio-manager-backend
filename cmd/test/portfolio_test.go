package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPortfolio_GetOwn tests getting user's own portfolios
func TestPortfolio_GetOwn(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_WithPagination", func(t *testing.T) {
		// Create test portfolios
		CreateTestPortfolio(testDB.DB, userID)
		CreateTestPortfolio(testDB.DB, userID)

		resp := MakeRequest(t, "GET", "/api/portfolios/own?page=1&limit=10", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			assert.Contains(t, body, "page")
			assert.Contains(t, body, "limit")
		})

		// Clean up
		cleanDatabase(testDB.DB)
	})

	t.Run("Success_EmptyList", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		resp := MakeRequest(t, "GET", "/api/portfolios/own?page=1&limit=10", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.Equal(t, 0, len(data))
		})
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/portfolios/own", nil, "")
		assert.Equal(t, 401, resp.Code)
	})

	t.Run("Pagination_DefaultValues", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/portfolios/own", nil, token)
		assert.Equal(t, 200, resp.Code)
	})
}

// TestPortfolio_Create tests creating a new portfolio
func TestPortfolio_Create(t *testing.T) {
	token := GetTestAuthToken()

	t.Run("Success_WithAllFields", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		payload := map[string]interface{}{
			"title":       "My Portfolio",
			"description": "A great portfolio",
		}

		resp := MakeRequest(t, "POST", "/api/portfolios/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "My Portfolio", data["title"])
			assert.Equal(t, "A great portfolio", data["description"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_OnlyRequiredFields", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		payload := map[string]interface{}{
			"title": "Minimal Portfolio",
		}

		resp := MakeRequest(t, "POST", "/api/portfolios/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Minimal Portfolio", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_EmptyTitle", func(t *testing.T) {
		payload := map[string]interface{}{
			"title": "",
		}

		resp := MakeRequest(t, "POST", "/api/portfolios/own", payload, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("ValidationError_MissingTitle", func(t *testing.T) {
		payload := map[string]interface{}{
			"description": "No title provided",
		}

		resp := MakeRequest(t, "POST", "/api/portfolios/own", payload, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		payload := map[string]interface{}{
			"title": "Test Portfolio",
		}

		resp := MakeRequest(t, "POST", "/api/portfolios/own", payload, "")
		assert.Equal(t, 401, resp.Code)
	})
}

// TestPortfolio_GetByID tests getting a specific portfolio
func TestPortfolio_GetByID(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_OwnPortfolio", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Test Portfolio", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/portfolios/own/99999", nil, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), nil, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestPortfolio_Update tests updating a portfolio
func TestPortfolio_Update(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_UpdateTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title": "Updated Portfolio",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Updated Portfolio", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateDescription", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":       "Test Portfolio",
			"description": "Updated description",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), payload, token)

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

		payload := map[string]interface{}{
			"title": "",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		payload := map[string]interface{}{
			"title": "Updated Portfolio",
		}

		resp := MakeRequest(t, "PUT", "/api/portfolios/own/99999", payload, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title": "Updated Portfolio",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestPortfolio_Delete tests deleting a portfolio
func TestPortfolio_Delete(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_DeletePortfolio", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), nil, token)
		assert.Equal(t, 200, resp.Code)

		// Verify deletion
		getResp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), nil, token)
		assert.Equal(t, 404, getResp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "DELETE", "/api/portfolios/own/99999", nil, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), nil, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_DeleteTwice_ReturnsNotFound", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		// First delete
		resp1 := MakeRequest(t, "DELETE", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), nil, token)
		assert.Equal(t, 200, resp1.Code)

		// Second delete should return 404
		resp2 := MakeRequest(t, "DELETE", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), nil, token)
		assert.Equal(t, 404, resp2.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestPortfolio_GetPublic tests getting public portfolios
func TestPortfolio_GetPublic(t *testing.T) {
	userID := GetTestUserID()

	t.Run("Success_PublicPortfolio", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		// No auth token needed for public endpoint
		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/public/%d", portfolio.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Test Portfolio", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/portfolios/public/99999", nil, "")
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Success_NoAuthRequired", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/public/%d", portfolio.ID), nil, "")
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestPortfolio_DuplicateCheck tests duplicate title validation
func TestPortfolio_DuplicateCheck(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Error_CreateDuplicate", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create first portfolio
		CreateTestPortfolioWithTitle(testDB.DB, userID, "My Portfolio")

		// Try to create second portfolio with same title
		payload := map[string]interface{}{
			"title":       "My Portfolio",
			"description": "Another portfolio",
		}

		resp := MakeRequest(t, "POST", "/api/portfolios/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		AssertJSONResponse(t, resp, 400, func(body map[string]interface{}) {
			assert.Contains(t, body, "error")
			assert.Contains(t, body["error"], "already exists")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_CreateDifferentTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create first portfolio
		CreateTestPortfolioWithTitle(testDB.DB, userID, "My Portfolio")

		// Create second portfolio with different title - should succeed
		payload := map[string]interface{}{
			"title":       "My Other Portfolio",
			"description": "Another portfolio",
		}

		resp := MakeRequest(t, "POST", "/api/portfolios/own", payload, token)
		assert.Equal(t, 201, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Error_UpdateToDuplicate", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create two portfolios
		portfolio1 := CreateTestPortfolioWithTitle(testDB.DB, userID, "Portfolio One")
		portfolio2 := CreateTestPortfolioWithTitle(testDB.DB, userID, "Portfolio Two")

		// Try to update portfolio2 to have same title as portfolio1
		payload := map[string]interface{}{
			"title": portfolio1.Title,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/portfolios/own/%d", portfolio2.ID), payload, token)
		assert.Equal(t, 400, resp.Code)

		AssertJSONResponse(t, resp, 400, func(body map[string]interface{}) {
			assert.Contains(t, body, "error")
			assert.Contains(t, body["error"], "already exists")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateSameTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create portfolio
		portfolio := CreateTestPortfolioWithTitle(testDB.DB, userID, "My Portfolio")

		// Update with same title but different description - should succeed
		payload := map[string]interface{}{
			"title":       "My Portfolio",
			"description": "Updated description",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/portfolios/own/%d", portfolio.ID), payload, token)
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})
}
