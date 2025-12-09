package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationQuery_GetPageAndLimit(t *testing.T) {
	tests := []struct {
		name          string
		query         PaginationQuery
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "Default values when empty",
			query:         PaginationQuery{},
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "Default values when zero",
			query:         PaginationQuery{Page: 0, Limit: 0},
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "Valid page and limit",
			query:         PaginationQuery{Page: 2, Limit: 20},
			expectedPage:  2,
			expectedLimit: 20,
		},
		{
			name:          "Limit exceeds maximum",
			query:         PaginationQuery{Page: 1, Limit: 150},
			expectedPage:  1,
			expectedLimit: 10, // Should default to 10 when exceeds max
		},
		{
			name:          "Limit at maximum boundary",
			query:         PaginationQuery{Page: 1, Limit: 100},
			expectedPage:  1,
			expectedLimit: 100,
		},
		{
			name:          "Negative page defaults to 1",
			query:         PaginationQuery{Page: -5, Limit: 20},
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "Negative limit defaults to 10",
			query:         PaginationQuery{Page: 2, Limit: -10},
			expectedPage:  2,
			expectedLimit: 10,
		},
		{
			name:          "Page 1 with limit 1",
			query:         PaginationQuery{Page: 1, Limit: 1},
			expectedPage:  1,
			expectedLimit: 1,
		},
		{
			name:          "Large page number",
			query:         PaginationQuery{Page: 999, Limit: 50},
			expectedPage:  999,
			expectedLimit: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, limit := tt.query.GetPageAndLimit()
			assert.Equal(t, tt.expectedPage, page, "Page should match expected")
			assert.Equal(t, tt.expectedLimit, limit, "Limit should match expected")
		})
	}
}

func TestPaginationQuery_GetOffset(t *testing.T) {
	tests := []struct {
		name           string
		query          PaginationQuery
		expectedOffset int
	}{
		{
			name:           "Page 1 - offset 0",
			query:          PaginationQuery{Page: 1, Limit: 10},
			expectedOffset: 0,
		},
		{
			name:           "Page 2 - offset 10",
			query:          PaginationQuery{Page: 2, Limit: 10},
			expectedOffset: 10,
		},
		{
			name:           "Page 3 - offset 40",
			query:          PaginationQuery{Page: 3, Limit: 20},
			expectedOffset: 40,
		},
		{
			name:           "Page 5 with limit 25 - offset 100",
			query:          PaginationQuery{Page: 5, Limit: 25},
			expectedOffset: 100,
		},
		{
			name:           "Default page and limit - offset 0",
			query:          PaginationQuery{},
			expectedOffset: 0,
		},
		{
			name:           "Page 10 with limit 50 - offset 450",
			query:          PaginationQuery{Page: 10, Limit: 50},
			expectedOffset: 450,
		},
		{
			name:           "Zero values default - offset 0",
			query:          PaginationQuery{Page: 0, Limit: 0},
			expectedOffset: 0,
		},
		{
			name:           "Page 1 with limit 1 - offset 0",
			query:          PaginationQuery{Page: 1, Limit: 1},
			expectedOffset: 0,
		},
		{
			name:           "Page 2 with limit 1 - offset 1",
			query:          PaginationQuery{Page: 2, Limit: 1},
			expectedOffset: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := tt.query.GetOffset()
			assert.Equal(t, tt.expectedOffset, offset, "Offset should match expected")
		})
	}
}

func TestErrorResponse(t *testing.T) {
	t.Run("Create error response", func(t *testing.T) {
		errResp := ErrorResponse{
			Error: "Test error message",
		}
		assert.Equal(t, "Test error message", errResp.Error)
	})

	t.Run("Empty error response", func(t *testing.T) {
		errResp := ErrorResponse{}
		assert.Equal(t, "", errResp.Error)
	})
}

func TestSuccessResponse(t *testing.T) {
	t.Run("Success response with data", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		resp := SuccessResponse{
			Message: "Success",
			Data:    data,
		}
		assert.Equal(t, "Success", resp.Message)
		assert.Equal(t, data, resp.Data)
	})

	t.Run("Success response without data", func(t *testing.T) {
		resp := SuccessResponse{
			Message: "Success",
		}
		assert.Equal(t, "Success", resp.Message)
		assert.Nil(t, resp.Data)
	})
}

func TestPaginatedResponse(t *testing.T) {
	t.Run("Create paginated response", func(t *testing.T) {
		data := []string{"item1", "item2", "item3"}
		resp := PaginatedResponse{
			Data:    data,
			Page:    2,
			Limit:   10,
			Total:   50,
			Message: "Success",
		}

		assert.Equal(t, data, resp.Data)
		assert.Equal(t, 2, resp.Page)
		assert.Equal(t, 10, resp.Limit)
		assert.Equal(t, int64(50), resp.Total)
		assert.Equal(t, "Success", resp.Message)
	})

	t.Run("Paginated response with empty data", func(t *testing.T) {
		resp := PaginatedResponse{
			Data:    []string{},
			Page:    1,
			Limit:   10,
			Total:   0,
			Message: "No results",
		}

		assert.NotNil(t, resp.Data)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 10, resp.Limit)
		assert.Equal(t, int64(0), resp.Total)
	})
}
