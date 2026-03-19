package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vinay/splitwise-grpc/api-gateway/clients"
	"github.com/vinay/splitwise-grpc/api-gateway/handlers"
	"github.com/vinay/splitwise-grpc/api-gateway/middleware"
	"github.com/vinay/splitwise-grpc/pkg/config"
	"github.com/vinay/splitwise-grpc/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config", "base")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Starting API Gateway", "port", ":8080")

	// Get service addresses from environment or use defaults
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50051"
	}

	expenseServiceAddr := os.Getenv("EXPENSE_SERVICE_ADDR")
	if expenseServiceAddr == "" {
		expenseServiceAddr = "localhost:50052"
	}

	// Initialize gRPC clients
	userClient, err := clients.NewUserClient(userServiceAddr)
	if err != nil {
		logger.Error("Failed to create user client", "error", err)
		os.Exit(1)
	}
	defer userClient.Close()

	expenseClient, err := clients.NewExpenseClient(expenseServiceAddr)
	if err != nil {
		logger.Error("Failed to create expense client", "error", err)
		os.Exit(1)
	}
	defer expenseClient.Close()

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware (order matters!)
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))

	// Rate limiting (100 requests per minute per user/IP)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute, logger)
	router.Use(rateLimiter.Middleware())
	logger.Info("Rate limiting enabled", "rate", 100, "window", "1m")

	// Audit logging
	auditLogger := middleware.NewAuditLogger(logger)
	router.Use(auditLogger.Middleware())
	logger.Info("Audit logging enabled")

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api-gateway",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth handlers
		authHandler := handlers.NewAuthHandler(userClient, cfg, logger)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.GET("/me", middleware.AuthRequired(cfg), authHandler.GetMe)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(cfg))
		{
			// User handlers
			userHandler := handlers.NewUserHandler(userClient, logger)
			users := protected.Group("/users")
			{
				users.GET("", userHandler.ListUsers)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.POST("/:id/friends", userHandler.AddFriend)
				users.GET("/:id/friends", userHandler.GetFriends)
			}

			// Expense handlers
			expenseHandler := handlers.NewExpenseHandler(expenseClient, logger)
			expenses := protected.Group("/expenses")
			{
				expenses.GET("", expenseHandler.ListExpenses)
				expenses.POST("", expenseHandler.CreateExpense)
				expenses.GET("/:id", expenseHandler.GetExpense)
				expenses.GET("/net-balance", expenseHandler.GetNetBalance)
				expenses.PUT("/:id", expenseHandler.UpdateExpense)
				expenses.DELETE("/:id", expenseHandler.DeleteExpense)
				expenses.POST("/settle", expenseHandler.SettleBalance)
			}
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("API Gateway listening", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("API Gateway stopped")
}
