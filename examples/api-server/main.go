package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"go.rumenx.com/chess/api"
	"go.rumenx.com/chess/config"
)

func main() {
	// Create configuration
	cfg := config.Default()

	// Create API server
	server := api.NewServer(cfg)

	// Create Gin router
	r := gin.Default()

	// Setup routes
	server.SetupRoutes(r)

	// Start server
	addr := cfg.GetServerAddress()
	log.Printf("Starting chess API server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
