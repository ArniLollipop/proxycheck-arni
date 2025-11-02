package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Initialize a single database
	db, err := gorm.Open(sqlite.Open("database/proxy.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate all models
	if err := db.AutoMigrate(&Proxy{}, &Settings{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Initialize Settings from the single database
	settings := SettingsDefault(db)

	// Initialize Gin router
	router := gin.Default()

	// Create handler instance
	h := handler{
		db:       db,
		settings: settings,
	}

	// API routes for proxies
	proxyRoutes := router.Group("api/proxy")
	{
		proxyRoutes.GET("", h.ProxyList)
		proxyRoutes.POST("", h.CreateProxy)
		proxyRoutes.GET(":id/verify", h.Verify)
		proxyRoutes.DELETE(":id", h.Delete)
		proxyRoutes.POST("verify-batch", h.VerifyBatch)
	}

	// Export routes
	exportRoutes := router.Group("api/export")
	{
		exportRoutes.GET("/all", h.ExportAll)
		exportRoutes.GET("/selected", h.ExportSelected)
	}

	// Setup HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Run server in a goroutine
	go func() {
		log.Println("Server running on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
