package test

import (
	"testing"

	models2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/stretchr/testify/assert"
)

// TestUser_GetUserDataSummary tests getting a summary of user's data
func TestUser_GetUserDataSummary(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_WithComplexData", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create comprehensive test data
		portfolio1 := CreateTestPortfolio(testDB.DB, userID)
		portfolio2 := CreateTestPortfolio(testDB.DB, userID)

		category1 := CreateTestCategory(testDB.DB, portfolio1.ID, userID)
		category2 := CreateTestCategory(testDB.DB, portfolio2.ID, userID)

		CreateTestSection(testDB.DB, portfolio1.ID, userID)
		CreateTestSection(testDB.DB, portfolio2.ID, userID)

		CreateTestProject(testDB.DB, category1.ID, userID)
		CreateTestProject(testDB.DB, category2.ID, userID)

		resp := MakeRequest(t, "GET", "/api/users/me/summary", nil, token)

		// Use AssertSuccessResponse helper for coverage
		AssertSuccessResponse(t, resp, 200)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})

			assert.Equal(t, float64(2), data["portfolios"])
			assert.Equal(t, float64(2), data["categories"])
			assert.Equal(t, float64(2), data["sections"])
			assert.Equal(t, float64(2), data["projects"])
			assert.Equal(t, float64(8), data["totalItems"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_EmptyUser", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		resp := MakeRequest(t, "GET", "/api/users/me/summary", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Contains(t, body, "data")
			data := body["data"].(map[string]interface{})

			assert.Equal(t, float64(0), data["portfolios"])
			assert.Equal(t, float64(0), data["categories"])
			assert.Equal(t, float64(0), data["sections"])
			assert.Equal(t, float64(0), data["projects"])
			assert.Equal(t, float64(0), data["totalItems"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_SinglePortfolio", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create single portfolio with nested data
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		CreateTestProject(testDB.DB, category.ID, userID)
		CreateTestSection(testDB.DB, portfolio.ID, userID)

		resp := MakeRequest(t, "GET", "/api/users/me/summary", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			data := body["data"].(map[string]interface{})

			assert.Equal(t, float64(1), data["portfolios"])
			assert.Equal(t, float64(1), data["categories"])
			assert.Equal(t, float64(1), data["sections"])
			assert.Equal(t, float64(1), data["projects"])
			assert.Equal(t, float64(4), data["totalItems"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_MultipleProjectsPerCategory", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)

		// Create multiple projects in same category
		CreateTestProject(testDB.DB, category.ID, userID)
		CreateTestProject(testDB.DB, category.ID, userID)
		CreateTestProject(testDB.DB, category.ID, userID)

		resp := MakeRequest(t, "GET", "/api/users/me/summary", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			data := body["data"].(map[string]interface{})
			assert.Equal(t, float64(3), data["projects"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := MakeRequest(t, "GET", "/api/users/me/summary", nil, "")
		assert.Equal(t, 401, resp.Code)
	})
}

// TestUser_CleanupUserData tests GDPR-compliant user data cleanup
func TestUser_CleanupUserData(t *testing.T) {
	token := GetTestAuthToken()
	userID := GetTestUserID()

	t.Run("Success_VerifyCascadeDelete", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create comprehensive test data
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
		project := CreateTestProject(testDB.DB, category.ID, userID)
		CreateTestImage(testDB.DB, project.ID, "project", userID)

		// Get counts before deletion
		var portfolioCount, categoryCount, sectionCount, projectCount, imageCount int64
		testDB.DB.Model(&models2.Portfolio{}).Where("owner_id = ?", userID).Count(&portfolioCount)
		testDB.DB.Model(&models2.Category{}).Where("owner_id = ?", userID).Count(&categoryCount)
		testDB.DB.Model(&models2.Section{}).Where("owner_id = ?", userID).Count(&sectionCount)
		testDB.DB.Model(&models2.Project{}).Where("owner_id = ?", userID).Count(&projectCount)
		testDB.DB.Model(&models2.Image{}).Where("owner_id = ?", userID).Count(&imageCount)

		assert.Equal(t, int64(1), portfolioCount)
		assert.Equal(t, int64(1), categoryCount)
		assert.Equal(t, int64(0), sectionCount) // No section created in this test
		assert.Equal(t, int64(1), projectCount)
		assert.Equal(t, int64(1), imageCount)

		// Execute cleanup
		resp := MakeRequest(t, "DELETE", "/api/users/me/data", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Equal(t, float64(1), body["portfoliosDeleted"])
			assert.Contains(t, body, "message")
		})

		// Verify CASCADE deleted everything
		testDB.DB.Model(&models2.Portfolio{}).Where("owner_id = ?", userID).Count(&portfolioCount)
		testDB.DB.Model(&models2.Category{}).Where("owner_id = ?", userID).Count(&categoryCount)
		testDB.DB.Model(&models2.Section{}).Where("owner_id = ?", userID).Count(&sectionCount)
		testDB.DB.Model(&models2.Project{}).Where("owner_id = ?", userID).Count(&projectCount)
		testDB.DB.Model(&models2.Image{}).Where("owner_id = ?", userID).Count(&imageCount)

		assert.Equal(t, int64(0), portfolioCount, "Portfolios should be deleted")
		assert.Equal(t, int64(0), categoryCount, "Categories should be cascade deleted")
		assert.Equal(t, int64(0), sectionCount, "Sections should be cascade deleted")
		assert.Equal(t, int64(0), projectCount, "Projects should be cascade deleted")
		assert.Equal(t, int64(0), imageCount, "Images should be cascade deleted")

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_MultiplePortfolios", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create 3 portfolios with nested data
		for i := 0; i < 3; i++ {
			portfolio := CreateTestPortfolio(testDB.DB, userID)
			category := CreateTestCategory(testDB.DB, portfolio.ID, userID)
			CreateTestProject(testDB.DB, category.ID, userID)
			CreateTestSection(testDB.DB, portfolio.ID, userID)
		}

		// Verify data exists
		var portfolioCount int64
		testDB.DB.Model(&models2.Portfolio{}).Where("owner_id = ?", userID).Count(&portfolioCount)
		assert.Equal(t, int64(3), portfolioCount)

		resp := MakeRequest(t, "DELETE", "/api/users/me/data", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Equal(t, float64(3), body["portfoliosDeleted"])
		})

		// Verify all deleted
		testDB.DB.Model(&models2.Portfolio{}).Where("owner_id = ?", userID).Count(&portfolioCount)
		assert.Equal(t, int64(0), portfolioCount)

		// Verify all related data deleted
		var categoryCount, sectionCount, projectCount int64
		testDB.DB.Model(&models2.Category{}).Where("owner_id = ?", userID).Count(&categoryCount)
		testDB.DB.Model(&models2.Section{}).Where("owner_id = ?", userID).Count(&sectionCount)
		testDB.DB.Model(&models2.Project{}).Where("owner_id = ?", userID).Count(&projectCount)

		assert.Equal(t, int64(0), categoryCount)
		assert.Equal(t, int64(0), sectionCount)
		assert.Equal(t, int64(0), projectCount)

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_EmptyUser", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// User has no data - should succeed with 0 deleted
		resp := MakeRequest(t, "DELETE", "/api/users/me/data", nil, token)

		AssertJSONResponse(t, resp, 200, func(body map[string]interface{}) {
			assert.Equal(t, float64(0), body["portfoliosDeleted"])
		})

		cleanDatabase(testDB.DB)
	})

	t.Run("Success_WithSectionContent", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create data including section content
		portfolio := CreateTestPortfolio(testDB.DB, userID)
		section := CreateTestSection(testDB.DB, portfolio.ID, userID)
		CreateTestSectionContent(testDB.DB, section.ID, userID)
		CreateTestSectionContent(testDB.DB, section.ID, userID)

		// Verify section content exists
		var contentCount int64
		testDB.DB.Model(&models2.SectionContent{}).Where("owner_id = ?", userID).Count(&contentCount)
		assert.Equal(t, int64(2), contentCount)

		resp := MakeRequest(t, "DELETE", "/api/users/me/data", nil, token)
		assert.Equal(t, 200, resp.Code)

		// Verify section content was cascade deleted
		testDB.DB.Model(&models2.SectionContent{}).Where("owner_id = ?", userID).Count(&contentCount)
		assert.Equal(t, int64(0), contentCount, "Section content should be cascade deleted")

		cleanDatabase(testDB.DB)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := MakeRequest(t, "DELETE", "/api/users/me/data", nil, "")
		assert.Equal(t, 401, resp.Code)
	})

	t.Run("Success_DeleteDoesNotAffectOtherUsers", func(t *testing.T) {
		cleanDatabase(testDB.DB)

		// Create data for test user
		testPortfolio := CreateTestPortfolio(testDB.DB, userID)

		// Create data for different user
		differentUserID := "different-user-789"
		otherPortfolio := CreateTestPortfolio(testDB.DB, differentUserID)
		otherCategory := CreateTestCategory(testDB.DB, otherPortfolio.ID, differentUserID)

		// Delete test user's data
		resp := MakeRequest(t, "DELETE", "/api/users/me/data", nil, token)
		assert.Equal(t, 200, resp.Code)

		// Verify test user's data is deleted
		var testUserPortfolioCount int64
		testDB.DB.Model(&models2.Portfolio{}).Where("id = ?", testPortfolio.ID).Count(&testUserPortfolioCount)
		assert.Equal(t, int64(0), testUserPortfolioCount)

		// Verify other user's data is NOT deleted
		var otherUserPortfolioCount, otherUserCategoryCount int64
		testDB.DB.Model(&models2.Portfolio{}).Where("id = ?", otherPortfolio.ID).Count(&otherUserPortfolioCount)
		testDB.DB.Model(&models2.Category{}).Where("id = ?", otherCategory.ID).Count(&otherUserCategoryCount)

		assert.Equal(t, int64(1), otherUserPortfolioCount, "Other user's portfolio should NOT be deleted")
		assert.Equal(t, int64(1), otherUserCategoryCount, "Other user's category should NOT be deleted")

		cleanDatabase(testDB.DB)
	})
}
