package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/vinay/splitwise-grpc/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Config holds the server configuration
type Config struct {
	Port          string
	Logger        *slog.Logger
	AuthValidator *auth.Validator   // Optional: if nil, auth is disabled
	PublicMethods map[string]bool   // Methods that don't require auth
	RateLimiter   *auth.RateLimiter // Optional: if nil, rate limiting is disabled
	AuditLogger   *auth.AuditLogger // Optional: if nil, audit logging is disabled
	TLSCertFile   string            // Optional: path to TLS cert file
	TLSKeyFile    string            // Optional: path to TLS key file
}

// RegisterServicesFunc is a callback to register gRPC services
type RegisterServicesFunc func(*grpc.Server)

// Run initializes and starts a gRPC server with production hardening
func Run(cfg Config, registerServices RegisterServicesFunc) error {
	log := cfg.Logger

	// 1. Setup Interceptors
	interceptors := []grpc.UnaryServerInterceptor{
		loggingInterceptor(log),
	}

	// Add audit logging if enabled
	if cfg.AuditLogger != nil {
		interceptors = append(interceptors, cfg.AuditLogger.Interceptor())
	}

	// Add auth interceptor if enabled
	if cfg.AuthValidator != nil {
		interceptors = append(interceptors, auth.Interceptor(cfg.AuthValidator, log, cfg.PublicMethods))
	}

	// Add rate limiting if enabled
	if cfg.RateLimiter != nil {
		interceptors = append(interceptors, cfg.RateLimiter.Interceptor())
	}

	interceptors = append(interceptors, recoveryInterceptor(log))

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(chainUnaryInterceptors(interceptors...)),
	}

	// Add TLS if configured
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
		log.Info("TLS enabled", "cert", cfg.TLSCertFile)
	}

	// 2. Create Listener
	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", cfg.Port, err)
	}

	// 3. Create Server
	s := grpc.NewServer(opts...)

	// 4. Register Health Server
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// 5. Register Services (Callback)
	registerServices(s)

	// 6. Register Reflection (for debugging/tools)
	reflection.Register(s)

	// 7. Start Server in Goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Info("gRPC server starting", "port", cfg.Port)
		if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			errChan <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	// 8. Graceful Shutdown Handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info("Shutting down server gracefully...")

		// Create a timeout for graceful shutdown
		stopped := make(chan struct{})
		go func() {
			s.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
			log.Info("Server stopped gracefully")
		case <-time.After(30 * time.Second):
			log.Warn("Graceful shutdown timed out, forcing stop")
			s.Stop()
		}
		return nil
	case err := <-errChan:
		return err
	}
}

// loggingInterceptor logs request details
func loggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		attrs := []any{
			slog.String("method", info.FullMethod),
			slog.Duration("duration", duration),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			log.Error("gRPC Request Failed", attrs...)
		} else {
			log.Info("gRPC Request", attrs...)
		}

		return resp, err
	}
}

// recoveryInterceptor recovers from panics
func recoveryInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic recovered",
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildChain(interceptors[i], info, chain)
		}
		return chain(ctx, req)
	}
}

func buildChain(interceptor grpc.UnaryServerInterceptor, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptor(ctx, req, info, handler)
	}
}
