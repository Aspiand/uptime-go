package api

import (
	"io"
	"net/http"
	"time"
	"uptime-go/internal/configuration"
	"uptime-go/internal/models"

	"github.com/gin-gonic/gin"
)

type ReportQueryParams struct {
	WithStat bool `form:"with_stat"`
}

type HistoryReportQueryParams struct {
	URL   string `form:"url"`
	Limit int    `form:"limit"`
}

type MonitorDailyUptimeStats struct {
	Date             string  `json:"date"`
	UptimePercentage float64 `json:"uptime_percentage"`
	TotalChecks      int     `json:"total_checks"` // 0 means no data for that day
}

type MonitorWithDailyUptimeStats struct {
	models.Monitor
	DailyStats []MonitorDailyUptimeStats `json:"stats,omitempty"`
}

func (s *Server) HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "uptime-go",
	})
}

func (s *Server) UpdateConfigHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read request body", "error": err.Error()})
		return
	}

	if err := configuration.UpdateConfig(s.configPath, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update configuration", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully. Please restart the application to apply changes."})
}

func (s *Server) GetMonitoringReport(c *gin.Context) {
	var params ReportQueryParams

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid query parameters",
		})
		return
	}

	var urls []string
	for _, m := range configuration.Config.Monitor {
		urls = append(urls, m.URL)
	}

	monitors, err := s.db.GetMonitors(urls)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve monitors",
			"error":   err.Error(),
		})
		return
	}

	if params.WithStat {
		var monitorsWithStats []MonitorWithDailyUptimeStats

		now := time.Now().UTC()
		todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
		ninetyDaysAgo := time.Date(todayEnd.Year(), todayEnd.Month(), todayEnd.Day(), 0, 0, 0, 0, todayEnd.Location()).AddDate(0, 0, -89)

		monitorURLs := make([]string, 0, len(monitors))
		for _, m := range monitors {
			monitorURLs = append(monitorURLs, m.URL)
		}

		historiesByMonitorURL, err := s.db.GetMonitorHistoriesForURLsInDateRange(monitorURLs, ninetyDaysAgo, todayEnd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to retrieve monitor histories for stats calculation",
			})
			return
		}

		for _, monitor := range monitors {
			histories, _ := historiesByMonitorURL[monitor.URL] // Will be empty slice if no histories
			dailyStats := calculateUptimeStats(histories, ninetyDaysAgo, todayEnd)

			monitorsWithStats = append(monitorsWithStats, MonitorWithDailyUptimeStats{
				Monitor:    monitor,
				DailyStats: dailyStats,
			})
		}
		c.JSON(http.StatusOK, monitorsWithStats)
		return
	}

	c.JSON(http.StatusOK, monitors)
}

func calculateUptimeStats(histories []models.MonitorHistory, from, to time.Time) []MonitorDailyUptimeStats {
	dailyStatsMap := make(map[string]struct {
		UpCount    int
		TotalCount int
	})

	// Initialize daily stats for the entire range
	for d := from; !d.After(to); d = d.Add(24 * time.Hour) {
		dateStr := d.Format("2006-01-02")
		dailyStatsMap[dateStr] = struct {
			UpCount    int
			TotalCount int
		}{UpCount: 0, TotalCount: 0}
	}

	for _, history := range histories {
		dateStr := history.CreatedAt.Format("2006-01-02")
		stats := dailyStatsMap[dateStr]
		stats.TotalCount++
		if history.IsUp {
			stats.UpCount++
		}
		dailyStatsMap[dateStr] = stats
	}

	var result []MonitorDailyUptimeStats
	for d := from; !d.After(to); d = d.Add(24 * time.Hour) {
		dateStr := d.Format("2006-01-02")
		stats := dailyStatsMap[dateStr]
		uptimePercentage := 0.0
		if stats.TotalCount > 0 {
			uptimePercentage = (float64(stats.UpCount) / float64(stats.TotalCount)) * 100
		}
		result = append(result, MonitorDailyUptimeStats{
			Date:             dateStr,
			UptimePercentage: uptimePercentage,
			TotalChecks:      stats.TotalCount,
		})
	}

	// If no histories at all, return nil to omit "stats" field in JSON response
	if len(histories) == 0 {
		return nil
	}

	return result
}

// func (s *Server) GetMonitoringReport(c *gin.Context) {
// 	var queryParams ReportQueryParams

// 	if err := c.ShouldBindQuery(&queryParams); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"message": "Invalid query parameters",
// 			"error":   err.Error(),
// 		})
// 		return
// 	}

// 	queryParams.URL = helper.NormalizeURL(queryParams.URL)

// 	if queryParams.Limit == 0 {
// 		queryParams.Limit = 1000
// 	}

// 	if queryParams.URL != "" {
// 		monitor, err := s.db.GetMonitorWithHistories(queryParams.URL, queryParams.Limit)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"message": "Failed to retrieve monitor details",
// 			})
// 			return
// 		}

// 		if monitor == nil {
// 			c.JSON(http.StatusNotFound, gin.H{"message": "Record not found"})
// 			return
// 		}

// 		// Handle single monitor with stats
// 		if queryParams.WithStat {
// 			now := time.Now().UTC()
// 			todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
// 			ninetyDaysAgo := time.Date(todayEnd.Year(), todayEnd.Month(), todayEnd.Day(), 0, 0, 0, 0, todayEnd.Location()).AddDate(0, 0, -89)

// 			histories, err := s.db.GetMonitorHistoriesInDateRange(monitor.URL, ninetyDaysAgo, todayEnd)
// 			if err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{
// 					"message": "Failed to retrieve monitor histories for stats calculation",
// 					"error":   err.Error(),
// 				})
// 				return
// 			}
// 			stats := calculateUptimeStats(histories, ninetyDaysAgo, todayEnd)
// 			monitorWithStats := MonitorWithStats{
// 				Monitor: *monitor,
// 				Stats:   stats,
// 			}
// 			c.JSON(http.StatusOK, monitorWithStats)
// 			return
// 		}

// 		c.JSON(http.StatusOK, monitor)
// 		return
// 	}

// 	// uptime% = (total_time - downtime) / total_time * 100

// 	var urls []string
// 	for _, m := range configuration.Config.Monitor {
// 		urls = append(urls, helper.NormalizeURL(m.URL))
// 	}

// 	monitors, err := s.db.GetMonitors(urls, queryParams.WithStat)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"message": "Failed to retrieve monitors",
// 		})
// 		return
// 	}

// 	if queryParams.WithStat {
// 		var monitorsWithStats []MonitorWithStats
// 		now := time.Now().UTC()
// 		todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
// 		ninetyDaysAgo := time.Date(todayEnd.Year(), todayEnd.Month(), todayEnd.Day(), 0, 0, 0, 0, todayEnd.Location()).AddDate(0, 0, -89)

// 		// Collect all monitor URLs for a single query
// 		monitorURLs := make([]string, 0, len(monitors))
// 		for _, m := range monitors {
// 			monitorURLs = append(monitorURLs, m.URL)
// 		}

// 		// Fetch all histories in a single query
// 		historiesByMonitorURL, err := s.db.GetMonitorHistoriesForURLsInDateRange(monitorURLs, ninetyDaysAgo, todayEnd)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"message": "Failed to retrieve monitor histories for stats calculation",
// 				"error":   err.Error(),
// 			})
// 			return
// 		}

// 		for _, monitor := range monitors {
// 			histories, _ := historiesByMonitorURL[monitor.URL] // Will be empty slice if no histories
// 			stats := calculateUptimeStats(histories, ninetyDaysAgo, todayEnd)
// 			monitorsWithStats = append(monitorsWithStats, MonitorWithStats{
// 				Monitor: monitor,
// 				Stats:   stats,
// 			})
// 		}
// 		c.JSON(http.StatusOK, monitorsWithStats)
// 		return
// 	}

// 	c.JSON(http.StatusOK, monitors)
// }

// func calculateUptimeStats(histories []models.MonitorHistory, from, to time.Time) []MonitorStats {
// 	dailyStats := make(map[string]struct {
// 		UpCount    int
// 		TotalCount int
// 	})

// 	for _, history := range histories {
// 		date := history.CreatedAt.Format("2006-01-02")
// 		stats := dailyStats[date]
// 		stats.TotalCount++
// 		if history.IsUp {
// 			stats.UpCount++
// 		}
// 		dailyStats[date] = stats
// 	}

// 	var result []MonitorStats
// 	for d := from; !d.After(to); d = d.Add(24 * time.Hour) {
// 		dateStr := d.Format("2006-01-02")
// 		stats, ok := dailyStats[dateStr]
// 		uptimePercentage := 0.0
// 		if ok && stats.TotalCount > 0 {
// 			uptimePercentage = (float64(stats.UpCount) / float64(stats.TotalCount)) * 100
// 		}
// 		result = append(result, MonitorStats{
// 			Date:             dateStr,
// 			UptimePercentage: uptimePercentage,
// 		})
// 	}
// 	return result
// }
