package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PortfolioHandler struct {
}

func NewPortfolioHandler(db *gorm.DB) *PortfolioHandler {
	return &PortfolioHandler{}
}

func (r *PortfolioHandler) GetByUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"portfolios": []gin.H{},
		"message":    "Success",
	})
}
