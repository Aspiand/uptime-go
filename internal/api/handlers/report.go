package handlers

import (
	"net/http"
	"uptime-go/internal/net/database"

	"github.com/gin-gonic/gin"
)

// ReportQueryParams struct to hold query parameters for the report endpoint
type ReportQueryParams struct {
	URL   string `form:"url"`
	Limit int    `form:"limit"` // gin's ShouldBindQuery doesn't natively support default tag, handle manually
}

func GetMonitoringReport(c *gin.Context) {
	var queryParams ReportQueryParams

	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid query parameters", "error": err.Error()})
		return
	}

	// Apply default for limit if not provided or invalid
	if queryParams.Limit == 0 { // If limit was not provided or parsed to 0, use default
		queryParams.Limit = 1000
	}

	db := database.Get()

	if queryParams.URL == "" {
		monitors, err := db.GetAllMonitors()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve monitors", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, monitors)
		return
	}

	monitor, err := db.GetMonitorWithHistories(queryParams.URL, queryParams.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve monitor details", "error": err.Error()})
		return
	}

	if monitor == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Record not found"})
		return
	}

	c.JSON(http.StatusOK, monitor)
}
