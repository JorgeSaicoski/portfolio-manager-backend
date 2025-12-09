package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// getBaseURL returns the base URL for test requests
func getBaseURL() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}
	return fmt.Sprintf("http://localhost:%s", port)
}

// MakeRequest creates and executes an HTTP request
func MakeRequest(t *testing.T, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		assert.NoError(t, err)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s%s", getBaseURL(), path)
	req, err := http.NewRequest(method, url, bodyReader)
	assert.NoError(t, err)

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Convert to httptest.ResponseRecorder for compatibility
	recorder := httptest.NewRecorder()
	bodyBytes, _ := io.ReadAll(resp.Body)
	recorder.Body.Write(bodyBytes)
	recorder.Code = resp.StatusCode
	for key, values := range resp.Header {
		for _, value := range values {
			recorder.Header().Add(key, value)
		}
	}

	return recorder
}

// AssertJSONResponse checks if response matches expected JSON
func AssertJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int, checkFunc func(map[string]interface{})) {
	assert.Equal(t, expectedCode, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	if checkFunc != nil {
		checkFunc(response)
	}
}

// AssertErrorResponse checks if response contains an error with expected message
func AssertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int, expectedErrorMsg string) {
	assert.Equal(t, expectedCode, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	if expectedErrorMsg != "" {
		assert.Contains(t, response, "error")
	}
}

// GetTestAuthToken returns a test JWT token
// In testing mode (TESTING_MODE=true), the auth middleware accepts
// any token and returns a test user with ID: test-user-123
func GetTestAuthToken() string {
	return "test-token-123"
}

// GetTestUserID returns a test user ID for creating fixtures
// This should match the user ID from the auth middleware test user (test-user-123)
func GetTestUserID() string {
	return "test-user-123" // Matches test user ID from auth middleware
}

// ParseJSONBody parses the response body into a map
func ParseJSONBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	return response
}

// AssertSuccessResponse checks for successful response with data
func AssertSuccessResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int) {
	assert.Equal(t, expectedCode, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Most success responses should contain a "data" field
	if expectedCode >= 200 && expectedCode < 300 {
		assert.Contains(t, response, "data")
	}
}

// AssertPaginatedResponse checks for paginated response structure
func AssertPaginatedResponse(t *testing.T, recorder *httptest.ResponseRecorder) {
	assert.Equal(t, 200, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "page")
	assert.Contains(t, response, "limit")
	assert.Contains(t, response, "total")
}
