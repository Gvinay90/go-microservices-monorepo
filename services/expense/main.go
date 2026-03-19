package main

import (
	"log"
	"os"

	"github.com/vinay/splitwise-grpc/internal/expense/handler"
	"github.com/vinay/splitwise-grpc/internal/expense/repository"
	"github.com/vinay/splitwise-grpc/internal/expense/service"
	"github.com/vinay/splitwise-grpc/pkg/auth"
	"github.com/vinay/splitwise-grpc/pkg/config"
	"github.com/vinay/splitwise-grpc/pkg/logger"
	"github.com/vinay/splitwise-grpc/pkg/server"
	pb "github.com/vinay/splitwise-grpc/proto/expense"
	"github.com/vinay/splitwise-grpc/storage"
	"google.golang.org/grpc"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load("config", "base")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Logger
	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	// 3. Initialize Database
	db, err := storage.New(cfg.Database)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// 4. Run Migrations
	log.Info("Running database migrations...")
	if err := db.AutoMigrate(&storage.Expense{}, &storage.Split{}, &storage.Settlement{}); err != nil {
		log.Error("Migration failed", "error", err)
		os.Exit(1)
	}

	// 5. Wire Clean Architecture Layers
	// Repository Layer (Data Access)
	expenseRepo := repository.NewGormRepository(db)

	// Service Layer (Business Logic)
	expenseService := service.NewExpenseService(expenseRepo, log)

	// Handler Layer (Transport)
	expenseHandler := handler.NewGRPCHandler(expenseService, log)

	// 6. Setup Authentication (if enabled)
	// Note: API Gateway handles full JWT validation
	// Internal services only extract claims from forwarded tokens
	var authValidator *auth.Validator
	var publicMethods map[string]bool
	if cfg.Auth.Enabled {
		authValidator = auth.NewValidator(cfg.Auth.KeycloakURL, cfg.Auth.Realm)
		// All expense methods require authentication
		publicMethods = map[string]bool{}
		log.Info("Authentication enabled (lightweight - trusting API Gateway)", "keycloak_url", cfg.Auth.KeycloakURL, "realm", cfg.Auth.Realm)
	} else {
		log.Warn("Authentication is DISABLED - all requests will be allowed")
	}

	// 7. Start Server using pkg/server
	// Note: Rate limiting and audit logging are handled by API Gateway
	srvCfg := server.Config{
		Port:          cfg.Server.Port,
		Logger:        log,
		AuthValidator: authValidator,
		PublicMethods: publicMethods,
		// RateLimiter and AuditLogger removed - handled by API Gateway
	}

	if err := server.Run(srvCfg, func(s *grpc.Server) {
		pb.RegisterExpenseServiceServer(s, expenseHandler)
	}); err != nil {
		log.Error("Server exited with error", "error", err)
		os.Exit(1)
	}
}
