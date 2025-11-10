package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"uptime-go/internal/api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Server struct {
	router     *gin.Engine
	HTTPServer *http.Server
}

type ServerConfig struct {
	Bind string
	Port string
}

func NewServer(cfg ServerConfig) *Server {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(accessLogger())

	server := &Server{
		router: router,
		HTTPServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%s", cfg.Bind, cfg.Port),
			Handler:      router.Handler(),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	server.setupRoutes()

	return server
}

func (s *Server) Start() error {
	log.Info().Str("address", s.HTTPServer.Addr).Msg("Starting api server")

	if err := s.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start api server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown() {
	log.Info().Msg("Stopping API server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := s.HTTPServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Error stopping API server")
	}

	log.Info().Msg("API server stopped successfully")
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "uptime-go",
		})
	})

	api := s.router.Group("/api/uptime-go")
	api.GET("/config")
	api.POST("/config")

	reportGroup := api.Group("/reports")
	reportGroup.GET("", handlers.GetMonitoringReport)
}

func accessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		logEvent := log.Info()
		if statusCode >= 400 {
			logEvent = log.Error()
		}

		logEvent.Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Str("latency", latency.String()).
			Msg("API request")
	}
}
