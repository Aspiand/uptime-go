package handlers

import (
	"uptime-go/internal/models"
	"uptime-go/internal/net/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetMonitoringReport(c *gin.Context) {
	db := database.Get()
	domainURL := c.Query("url")

	if domainURL == "" {
		var monitors []models.Monitor
		db.DB.Find(&monitors)
		c.JSON(200, monitors)
		return
	}

	var monitor models.Monitor
	db.DB.
		Preload("Histories", func(db *gorm.DB) *gorm.DB {
			return db.Order("monitor_histories.created_at DESC").Limit(100)
		}).
		Where("url = ?", domainURL).
		Find(&monitor)

	if monitor.IsNotExists() {
		c.JSON(404, gin.H{"message": "Record not found"})
		return
	}

	c.JSON(200, monitor)
}
