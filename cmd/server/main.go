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
	handler "url-shortener1/internal/handlers"
	"url-shortener1/internal/repository"
	"url-shortener1/internal/service"
	"url-shortener1/internal/storage/sqlite"
)

func main() {
	// Initialize database
	storage, err := sqlite.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repository
	repo := repository.NewSQLiteRepository(storage)
	defer func() {
		if err := storage.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize service
	urlService := service.NewURLService(repo, 6)

	// Initialize handler
	h := handler.NewHandler(urlService)

	// Setup router
	r := gin.Default()

	// Routes
	r.POST("/api/shorten", h.ShortenURL)
	r.GET("/:short", h.Redirect)
	r.GET("/api/stat/:short", h.Stat)

	// Create server
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
