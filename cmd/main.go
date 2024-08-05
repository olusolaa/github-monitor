package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/olusolaa/github-monitor/config"
	httpHandlers "github.com/olusolaa/github-monitor/internal/adapters/http"
	"github.com/olusolaa/github-monitor/internal/container"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	logger.InitLogger()

	// Create DI container and initialize components
	diContainer := container.NewContainer(cfg)
	defer diContainer.Close() // Ensure resources are properly closed

	// Fetch initial repository information and publish an event
	diContainer.InitializeRepository()

	// Start message consumers for commit processing and monitoring
	go diContainer.StartServices()

	// Set up the HTTP router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Register routes with the HTTP router
	httpHandlers.RegisterRoutes(r, diContainer.GetRepoService(), diContainer.GetCommitService())

	// Define and start the HTTP server
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	// Run the server in a goroutine
	go func() {
		log.Println("Server starting on " + cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError(err)
		}
	}()

	// Handle graceful shutdown
	gracefulShutdown(server)
}

func gracefulShutdown(server *http.Server) {
	// Set up channel to listen for termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create a deadline for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		logger.LogError(err)
	}

	log.Println("Server exiting")
}
