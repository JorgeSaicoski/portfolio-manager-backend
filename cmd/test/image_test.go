package test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestImage_Upload tests image upload
func TestImage_Upload(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_UploadImage", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		// Create a test image file
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add form fields
		_ = writer.WriteField("entity_type", "project")
		_ = writer.WriteField("entity_id", fmt.Sprintf("%d", project.ID))
		_ = writer.WriteField("type", "image")
		_ = writer.WriteField("alt", "Test image")
		_ = writer.WriteField("is_main", "true")

		// Create a simple test image (1x1 PNG)
		testImageData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
			0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
			0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
			0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
			0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		}

		part, _ := writer.CreateFormFile("file", "test.png")
		_, _ = part.Write(testImageData)

		writer.Close()

		// Make request
		req, _ := http.NewRequest("POST", "/api/images/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)

		resp := ExecuteRequest(req)

		AssertJSONResponse(t, resp, 201, func(responseBody map[string]interface{}) {
			assert.Contains(t, responseBody, "data")
			data := responseBody["data"].(map[string]interface{})
			assert.Equal(t, "Test image", data["alt"])
			assert.Equal(t, "project", data["entity_type"])
			assert.Equal(t, true, data["is_main"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Fail_NoFile", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("entity_type", "project")
		_ = writer.WriteField("entity_id", fmt.Sprintf("%d", project.ID))
		_ = writer.WriteField("type", "image")
		writer.Close()

		req, _ := http.NewRequest("POST", "/api/images/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)

		resp := ExecuteRequest(req)
		assert.Equal(t, 400, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req, _ := http.NewRequest("POST", "/api/images/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp := ExecuteRequest(req)
		assert.Equal(t, 401, resp.Code)
	})
}

// TestImage_GetByEntity tests getting images for an entity
func TestImage_GetByEntity(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_GetImages", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)

		// Create test images directly in DB
		CreateTestImage(testDB.DB, project.ID, "project", userID)
		CreateTestImage(testDB.DB, project.ID, "project", userID)

		url := fmt.Sprintf("/api/images?entity_type=project&entity_id=%d", project.ID)
		resp := MakeRequest(t, "GET", url, nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].([]interface{})
			assert.Equal(t, 2, len(data))
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Fail_MissingParameters", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/images", nil, token)
		assert.Equal(t, 400, resp.Code)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/images?entity_type=project&entity_id=1", nil, "")
		assert.Equal(t, 401, resp.Code)
	})
}

// TestImage_Delete tests deleting an image
func TestImage_Delete(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_DeleteImage", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)
		image := CreateTestImage(testDB.DB, project.ID, "project", userID)

		url := fmt.Sprintf("/api/images/%d", image.ID)
		resp := MakeRequest(t, "DELETE", url, nil, token)

		assert.Equal(t, 200, resp.Code)

		cleanDatabase(testDB.DB)
	})

	t.Run("Fail_NotFound", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		resp := MakeRequest(t, "DELETE", "/api/images/99999", nil, token)
		assert.Equal(t, 403, resp.Code) // Forbidden because ownership check fails first

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := MakeRequest(t, "DELETE", "/api/images/1", nil, "")
		assert.Equal(t, 401, resp.Code)
	})
}

// TestImage_Update tests updating image metadata
func TestImage_Update(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_UpdateAlt", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)
		image := CreateTestImage(testDB.DB, project.ID, "project", userID)

		payload := map[string]interface{}{
			"alt": "Updated alt text",
		}

		url := fmt.Sprintf("/api/images/%d", image.ID)
		resp := MakeRequest(t, "PUT", url, payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Updated alt text", data["alt"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_UpdateIsMain", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)
		image := CreateTestImage(testDB.DB, project.ID, "project", userID)

		payload := map[string]interface{}{
			"is_main": true,
		}

		url := fmt.Sprintf("/api/images/%d", image.ID)
		resp := MakeRequest(t, "PUT", url, payload, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})
			assert.Equal(t, true, data["is_main"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		payload := map[string]interface{}{"alt": "Test"}
		resp := MakeRequest(t, "PUT", "/api/images/1", payload, "")
		assert.Equal(t, 401, resp.Code)
	})
}
