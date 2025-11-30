package test

import (
	"fmt"
	"testing"

	models2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/stretchr/testify/assert"
)

// TestSectionContent_Create tests creating a new section content block
func TestSectionContent_Create(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_WithTextContent", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Setup: portfolio -> section
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"section_id": section.ID,
			"type":       "text",
			"content":    "Sample text content for testing",
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "text", data["type"])
			assert.Equal(t, "Sample text content for testing", data["content"])
			assert.Equal(t, float64(0), data["order"]) // Default order
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_WithImageContent", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"section_id": section.ID,
			"type":       "image",
			"content":    "Image description text",
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "image", data["type"])
			assert.Equal(t, "Image description text", data["content"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_WithCustomOrder", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"section_id": section.ID,
			"type":       "text",
			"content":    "Content with custom order",
			"order":      5,
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			data := body["data"].(map[string]interface{})
			assert.Equal(t, float64(5), data["order"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_WithMetadata", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		metadata := `{"key": "value", "number": 42}`
		payload := map[string]interface{}{
			"section_id": section.ID,
			"type":       "text",
			"content":    "Content with metadata",
			"metadata":   metadata,
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)

		AssertJSONResponse(t, resp, 201, func(body map[string]interface{}) {
			data := body["data"].(map[string]interface{})
			assert.Contains(t, data, "metadata")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_MissingType", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"section_id": section.ID,
			"content":    "Content without type",
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_InvalidContentType", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"section_id": section.ID,
			"type":       "invalid_type",
			"content":    "Content",
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_MissingContent", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"section_id": section.ID,
			"type":       "text",
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_MissingSectionID", func(t *testing.T) {
		payload := map[string]interface{}{
			"type":    "text",
			"content": "Content without section",
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("NotFound_InvalidSectionID", func(t *testing.T) {
		payload := map[string]interface{}{
			"section_id": 99999,
			"type":       "text",
			"content":    "Content for non-existent section",
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Forbidden_PortfolioOwnershipCheck", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create section owned by different user
		differentUserID := "different-user-456"
		portfolio := CreateTestPortfolio(testDB.DB, differentUserID)
		section := CreateTestSection(testDB.DB, portfolio.ID, differentUserID)

		// Try to create content as test user
		payload := map[string]interface{}{
			"section_id": section.ID,
			"type":       "text",
			"content":    "Unauthorized content",
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, token)
		assert.Equal(t, 403, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		payload := map[string]interface{}{
			"section_id": section.ID,
			"type":       "text",
			"content":    "Content",
		}

		resp := MakeRequest(t, "POST", "/api/section-contents/own", payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestSectionContent_GetBySectionID tests retrieving all content blocks for a section
func TestSectionContent_GetBySectionID(t *testing.T) {
	userID := GetTestUserID()

	t.Run("Success_MultipleContents_OrderedCorrectly", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		// Create contents with different orders
		CreateTestSectionContentWithOrder(testDB.DB, section.ID, userID, 3)
		CreateTestSectionContentWithOrder(testDB.DB, section.ID, userID, 1)
		CreateTestSectionContentWithOrder(testDB.DB, section.ID, userID, 2)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/sections/%d/contents", section.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.Equal(t, 3, len(data))

			// Verify ordering
			assert.Equal(t, float64(1), data[0].(map[string]interface{})["order"])
			assert.Equal(t, float64(2), data[1].(map[string]interface{})["order"])
			assert.Equal(t, float64(3), data[2].(map[string]interface{})["order"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_EmptyList", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/sections/%d/contents", section.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			data := body["data"].([]interface{})
			assert.Equal(t, 0, len(data))
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_NoAuthRequired", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		CreateTestSectionContent(testDB.DB, section.ID, userID)

		// Public endpoint - no token needed
		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/sections/%d/contents", section.ID), nil, "")
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("BadRequest_InvalidSectionID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/sections/invalid/contents", nil, "")
		assert.Equal(t, 400, resp.Code)
	})
}

// TestSectionContent_GetByID tests retrieving a single content block
func TestSectionContent_GetByID(t *testing.T) {
	userID := GetTestUserID()

	t.Run("Success_ValidID", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/section-contents/%d", content.ID), nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Test content", data["content"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/section-contents/99999", nil, "")
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("BadRequest_InvalidIDFormat", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/section-contents/invalid", nil, "")
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("Success_NoAuthRequired", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		// Public endpoint - no token needed
		resp := MakeRequest(t, "GET", fmt.Sprintf("/api/section-contents/%d", content.ID), nil, "")
		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestSectionContent_Update tests updating a section content block
func TestSectionContent_Update(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_UpdateContent", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		payload := map[string]interface{}{
			"content": "Updated content text",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/section-contents/own/%d", content.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Updated content text", data["content"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateType", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		payload := map[string]interface{}{
			"type": "image",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/section-contents/own/%d", content.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "image", data["type"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateMetadata", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		metadata := `{"updated": true}`
		payload := map[string]interface{}{
			"metadata": metadata,
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/section-contents/own/%d", content.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		payload := map[string]interface{}{
			"content": "Updated content",
		}

		resp := MakeRequest(t, "PUT", "/api/section-contents/own/99999", payload, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("BadRequest_InvalidIDFormat", func(t *testing.T) {
		payload := map[string]interface{}{
			"content": "Updated content",
		}

		resp := MakeRequest(t, "PUT", "/api/section-contents/own/invalid", payload, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("Forbidden_OwnershipCheck", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create content owned by different user
		differentUserID := "different-user-456"
		portfolio := CreateTestPortfolio(testDB.DB, differentUserID)
		section := CreateTestSection(testDB.DB, portfolio.ID, differentUserID)
		content := CreateTestSectionContent(testDB.DB, section.ID, differentUserID)

		// Try to update as test user
		payload := map[string]interface{}{
			"content": "Unauthorized update",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/section-contents/own/%d", content.ID), payload, token)
		assert.Equal(t, 403, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		payload := map[string]interface{}{
			"content": "Updated content",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/section-contents/own/%d", content.ID), payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_InvalidType", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		payload := map[string]interface{}{
			"type": "invalid_type",
		}

		resp := MakeRequest(t, "PUT", fmt.Sprintf("/api/section-contents/own/%d", content.ID), payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestSectionContent_UpdateOrder tests updating just the order of a content block
func TestSectionContent_UpdateOrder(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_UpdateOrder", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContentWithOrder(testDB.DB, section.ID, userID, 1)

		payload := map[string]interface{}{
			"order": 5,
		}

		resp := MakeRequest(t, "PATCH", fmt.Sprintf("/api/section-contents/own/%d/order", content.ID), payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			data := body["data"].(map[string]interface{})
			assert.Equal(t, float64(5), data["order"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		payload := map[string]interface{}{
			"order": 5,
		}

		resp := MakeRequest(t, "PATCH", "/api/section-contents/own/99999/order", payload, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("Forbidden_OwnershipCheck", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create content owned by different user
		differentUserID := "different-user-456"
		portfolio := CreateTestPortfolio(testDB.DB, differentUserID)
		section := CreateTestSection(testDB.DB, portfolio.ID, differentUserID)
		content := CreateTestSectionContent(testDB.DB, section.ID, differentUserID)

		payload := map[string]interface{}{
			"order": 5,
		}

		resp := MakeRequest(t, "PATCH", fmt.Sprintf("/api/section-contents/own/%d/order", content.ID), payload, token)
		assert.Equal(t, 403, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		payload := map[string]interface{}{
			"order": 5,
		}

		resp := MakeRequest(t, "PATCH", fmt.Sprintf("/api/section-contents/own/%d/order", content.ID), payload, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("ValidationError_MissingOrder", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		payload := map[string]interface{}{}

		resp := MakeRequest(t, "PATCH", fmt.Sprintf("/api/section-contents/own/%d/order", content.ID), payload, token)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})
}

// TestSectionContent_Delete tests deleting a section content block
func TestSectionContent_Delete(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_DeleteContent", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/section-contents/own/%d", content.ID), nil, token)
		assert.Equal(t, 200, resp.Code)

		// Verify content is deleted
		var count int64
		testDB.DB.Model(&models2.SectionContent{}).Where("id = ?", content.ID).Count(&count)
		assert.Equal(t, int64(0), count)

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_WithImageMetadata", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)

		// Create content with image metadata
		content := CreateTestSectionContentWithImage(testDB.DB, section.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/section-contents/own/%d", content.ID), nil, token)
		assert.Equal(t, 200, resp.Code)

		// Verify content is deleted
		var count int64
		testDB.DB.Model(&models2.SectionContent{}).Where("id = ?", content.ID).Count(&count)
		assert.Equal(t, int64(0), count)

		cleanDatabase(testDB.DB)
	})

	t.Run("NotFound_InvalidID", func(t *testing.T) {
		resp := MakeRequest(t, "DELETE", "/api/section-contents/own/99999", nil, token)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("BadRequest_InvalidIDFormat", func(t *testing.T) {
		resp := MakeRequest(t, "DELETE", "/api/section-contents/own/invalid", nil, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("Forbidden_OwnershipCheck", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create content owned by different user
		differentUserID := "different-user-456"
		portfolio := CreateTestPortfolio(testDB.DB, differentUserID)
		section := CreateTestSection(testDB.DB, portfolio.ID, differentUserID)
		content := CreateTestSectionContent(testDB.DB, section.ID, differentUserID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/section-contents/own/%d", content.ID), nil, token)
		assert.Equal(t, 403, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		resp := MakeRequest(t, "DELETE", fmt.Sprintf("/api/section-contents/own/%d", content.ID), nil, "")
		assert.Equal(t, 401, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_DeleteTwice_ReturnsNotFound", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		content := CreateTestSectionContent(testDB.DB, section.ID, userID)

		// First delete
		resp1 := MakeRequest(t, "DELETE", fmt.Sprintf("/api/section-contents/own/%d", content.ID), nil, token)
		assert.Equal(t, 200, resp1.Code)

		// Second delete should return 404
		resp2 := MakeRequest(t, "DELETE", fmt.Sprintf("/api/section-contents/own/%d", content.ID), nil, token)
		assert.Equal(t, 404, resp2.Code)

		cleanDatabase(testDB.DB)
	})
}
