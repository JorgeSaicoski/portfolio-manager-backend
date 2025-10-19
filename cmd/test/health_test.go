package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth_Basic(t *testing.T) {
	t.Run("Health_Check_Success", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/health", nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "status")
			assert.Equal(t, "healthy", body["status"])
		})
	})

	t.Run("Health_Check_Database_Connected", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/health", nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "database")
			assert.Equal(t, "connected", body["database"])
		})
	})
}

func TestHealth_Response_Format(t *testing.T) {
	t.Run("Has_Timestamp", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/health", nil, "")

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "timestamp")
		})
	})

	t.Run("JSON_Content_Type", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/health", nil, "")

		assert.Equal(t, 200, resp.Code)
		contentType := resp.Header().Get("Content-Type")
		assert.Contains(t, contentType, "application/json")
	})
}

func TestHealth_Availability(t *testing.T) {
	t.Run("Multiple_Requests_Succeed", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			resp := MakeRequest(t, "GET", "/health", nil, "")
			assert.Equal(t, 200, resp.Code)
		}
	})

	t.Run("No_Authentication_Required", func(t *testing.T) {
		// Health endpoint should be accessible without auth
		resp := MakeRequest(t, "GET", "/health", nil, "")
		assert.Equal(t, 200, resp.Code)
	})
}
