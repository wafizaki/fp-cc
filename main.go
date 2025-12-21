package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-manager/config"
	"tenant-manager/database"
	"tenant-manager/handlers"
	"tenant-manager/services"
	"tenant-manager/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	if err := database.InitDB(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	dockerClient, err := utils.NewDockerClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	tenantService := services.NewTenantService(dockerClient, cfg.BaseDir)

	tenantHandler := handlers.NewTenantHandler(tenantService)

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"message": "Server is healthy",
			"data": gin.H{
				"status": "ok",
			},
		})
	})

	api := router.Group("/api")
	{
		tenants := api.Group("/tenants")
		{
			tenants.POST("", tenantHandler.CreateTenant)
			tenants.GET("", tenantHandler.ListTenants)
			tenants.GET("/:name", tenantHandler.GetTenant)
			tenants.DELETE("/:name", tenantHandler.DeleteTenant)

			tenants.PUT("/:name/stop", tenantHandler.StopContainer)
			tenants.PUT("/:name/start", tenantHandler.StartContainer)
		}
	}

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting Tenant Management API server on %s", addr)
	log.Printf("Database: %s", cfg.DBPath)
	log.Printf("Base directory: %s", cfg.BaseDir)
	log.Println("API endpoints:")
	log.Println("  GET    /health")
	log.Println("  POST   /api/tenants")
	log.Println("  GET    /api/tenants")
	log.Println("  GET    /api/tenants/:name")
	log.Println("  DELETE /api/tenants/:name")
	log.Println("  PUT    /api/tenants/:name/stop")
	log.Println("  PUT    /api/tenants/:name/start")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

