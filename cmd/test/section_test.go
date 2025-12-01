package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSection_GetOwn tests getting user's own sections
func TestSection_GetOwn(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_WithPagination", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		CreateTestSection(testDB.DB, portfolio.ID, userID)
		CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", "/api/sections/own?page=1&limit=10", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			assert.Contains(t, body, "page")
			assert.Contains(t, body, "limit")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_EmptyList", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		resp := MakeRequest(t, "GET", "/api/sections/own?page=1&limit=10", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.Equal(t, 0, len(data))
		})
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/sections/own", nil, "")
		assert.Equal(t, 401, resp.Code)
	})

	t.Run("Pagination_LargePageNumber", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		CreateTestSection(testDB.DB, portfolio.ID, userID) // Create 1 item

		resp := MakeRequest(t, "GET", "/api/sections/own?page=100&limit=10", nil, token)
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

		resp := MakeRequest(t, "GET", "/api/sections/own?page=1&limit=101", nil, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Pagination_InvalidPageZero", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/sections/own?page=0&limit=10", nil, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("Pagination_InvalidNegativePage", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/sections/own?page=-1&limit=10", nil, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("Pagination_BoundaryExactlyAtLimit", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		// Create exactly 10 items
		for i := 0; i < 10; i++ {
			CreateTestSectionWithTitle(testDB.DB, portfolio.ID, userID, fmt.Sprintf("Section %d", i))
		}

		resp := MakeRequest(t, "GET", "/api/sections/own?page=1&limit=10", nil, token)
		AssertJSONResponse(t, resp, 200, func(data map[string]interface{}) {
			items := data["data"].([]interface{})
			assert.Equal(t, 10, len(items))
			assert.Equal(t, float64(1), data["total_pages"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Pagination_SecondPageWithRemainder", func(t *testing.T) {
		cleanDatabase(testDB.DB)
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		// Create 15 items (page 1: 10, page 2: 5)
		for i := 0; i < 15; i++ {
			CreateTestSectionWithTitle(testDB.DB, portfolio.ID, userID, fmt.Sprintf("Section %d", i))
		}

		resp := MakeRequest(t, "GET", "/api/sections/own?page=2&limit=10", nil, token)
		AssertJSONResponse(t, resp, 200, func(data map[string]interface{}) {
			items := data["data"].([]interface{})
			assert.Equal(t, 5, len(items))
			assert.Equal(t, float64(2), data["page"])
			assert.Equal(t, float64(2), data["total_pages"])
		})

		cleanDatabase(testDB.DB)
	})
}

// TestSection_Create tests creating a new section
func TestSection_Create(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_WithAllFields", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "About Me",
			"description":  "Introduction section",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/sections/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "About Me", data["title"])
			assert.Equal(t, "Introduction section", data["description"])
			assert.Equal(t, "text", data["type"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_OnlyRequiredFields", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "Skills",
			"type":         "list",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/sections/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Skills", data["title"])
			assert.Equal(t, "list", data["type"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_DifferentTypes", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		types := []string{"text", "image", "video", "list", "code"}

		for _, sectionType := range types {
			payload := map[string]interface{}{
				"title":        fmt.Sprintf("Section %s", sectionType),
				"type":         sectionType,
				"portfolio_id": portfolio.ID,
			}

			resp := MakeRequest(t, "POST", "/api/sections/own", payload, token)

			AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
				assert.Contains(t, body, "data")
				data := body["data"].(map[string]interface{})
				assert.Equal(t, sectionType, data["type"])
			})
		}

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_EmptyTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/sections/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_MissingType", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "Test Section",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/sections/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_MissingPortfolioID", func(t *testing.T) {
		payload := map[string]interface{}{
			"title": "Test Section",
			"type":  "text",
		}

		resp := MakeRequest(t, "POST", "/api/sections/own", payload, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("ValidationError_InvalidPortfolioID", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":        "Test Section",
			"type":         "text",
			"portfolio_id": 99999,
		}

		resp := MakeRequest(t, "POST", "/api/sections/own", payload, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "Test Section",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/sections/own", payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestSection_GetByID tests getting a specific section
func TestSection_GetByID(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_OwnSection", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/sections/own/%d", section.ID), nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Test Section", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/sections/own/99999", nil, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/sections/own/%d", section.ID), nil, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestSection_Update tests updating a section
func TestSection_Update(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_UpdateTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "Updated Section",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/sections/own/%d", section.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Updated Section", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateDescription", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "Test Section",
			"description":  "Updated description",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/sections/own/%d", section.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Updated description", data["description"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateType", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "Test Section",
			"type":         "image",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/sections/own/%d", section.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "image", data["type"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_EmptyTitle", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/sections/own/%d", section.ID), payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_EmptyType", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "Test Section",
			"type":         "",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/sections/own/%d", section.ID), payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		payload := map[string]interface{}{
			"title":        "Updated Section",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", "/api/sections/own/99999", payload, token)
		assert.Equal(t, 404, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"title":        "Updated Section",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/sections/own/%d", section.ID), payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestSection_Delete tests deleting a section
func TestSection_Delete(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_DeleteSection", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/sections/own/%d", section.ID), nil, token)
		assert.Equal(t, 200, resp.Code)

		// Verify deletion
		getResp := MakeRequest(t, "GET", fmt.Sprintf("/api/sections/own/%d", section.ID), nil, token)
		assert.Equal(t, 404, getResp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "DELETE", "/api/sections/own/99999", nil, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/sections/own/%d", section.ID), nil, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_DeleteMultipleSections", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section1 := CreateTestSection(testDB.DB, portfolio.ID, userID)
		section2 := CreateTestSection(testDB.DB, portfolio.ID, userID)

		// Delete first section
		resp1 := MakeRequest(t, "DELETE", fmt.Sprintf("/api/sections/own/%d", section1.ID), nil, token)
		assert.Equal(t, 200, resp1.Code)

		// Delete second section
		resp2 := MakeRequest(t, "DELETE", fmt.Sprintf("/api/sections/own/%d", section2.ID), nil, token)
		assert.Equal(t, 200, resp2.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestSection_GetPublic tests getting public sections
func TestSection_GetPublic(t *testing.T) {
	userID := GetTestUserID()

	t.Run("Success_PublicSection", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/sections/public/%d", section.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Test Section", data["title"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/sections/public/99999", nil, "")
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Success_NoAuthRequired", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/sections/public/%d", section.ID), nil, "")
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestSection_GetByPortfolio tests getting sections by portfolio
func TestSection_GetByPortfolio(t *testing.T) {
	userID := GetTestUserID()

	t.Run("Success_GetSectionsByPortfolio", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		CreateTestSection(testDB.DB, portfolio.ID, userID)
		CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/public/%d/sections", portfolio.ID), nil, "")

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

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/public/%d/sections", portfolio.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.Equal(t, 0, len(data))
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_FilterByType", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		CreateTestSection(testDB.DB, portfolio.ID, userID) // type: text

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/portfolios/public/%d/sections?type=text", portfolio.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.GreaterOrEqual(t, len(data), 1)
		})

		cleanDatabase(testDB.DB)
	})
}

// TestSection_DuplicateCheck tests duplicate title validation within same portfolio
func TestSection_DuplicateCheck(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Error_CreateDuplicate", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		// Create first section
		CreateTestSectionWithTitle(testDB.DB, portfolio.ID, userID, "About Me")

		// Try to create second section with same title in same portfolio
		payload := map[string]interface{}{
			"title":        "About Me",
			"description":  "Another about section",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "POST", "/api/sections/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		AssertJSONResponse(t, resp, 400, func(body map[string]interface{}) {
			assert.Contains(t, body, "error")
			assert.Contains(t, body["error"], "already exists")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_CreateSameTitleDifferentPortfolio", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio1 := CreateTestPortfolioWithTitle(testDB.DB, userID, "Portfolio 1")
		portfolio2 := CreateTestPortfolioWithTitle(testDB.DB, userID, "Portfolio 2")

		// Create section in first portfolio
		CreateTestSectionWithTitle(testDB.DB, portfolio1.ID, userID, "About Me")

		// Create section with same title but in different portfolio - should succeed
		payload := map[string]interface{}{
			"title":        "About Me",
			"description":  "About section for portfolio 2",
			"type":         "text",
			"portfolio_id": portfolio2.ID,
		}

		resp := MakeRequest(t, "POST", "/api/sections/own", payload, token)
		assert.Equal(t, 201, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Error_UpdateToDuplicate", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)

		// Create two sections
		section1 := CreateTestSectionWithTitle(testDB.DB, portfolio.ID, userID, "Section One")
		section2 := CreateTestSectionWithTitle(testDB.DB, portfolio.ID, userID, "Section Two")

		// Try to update section2 to have same title as section1
		payload := map[string]interface{}{
			"title":        section1.Title,
			"description":  "Updated description",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/sections/own/%d", section2.ID), payload, token)
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

		// Create section
		section := CreateTestSectionWithTitle(testDB.DB, portfolio.ID, userID, "My Section")

		// Update with same title but different description - should succeed
		payload := map[string]interface{}{
			"title":        "My Section",
			"description":  "Updated description",
			"type":         "text",
			"portfolio_id": portfolio.ID,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/sections/own/%d", section.ID), payload, token)
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})
}
