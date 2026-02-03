package auth

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuditLogger logs security-relevant events
type AuditLogger struct {
	log *slog.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(log *slog.Logger) *AuditLogger {
	return &AuditLogger{log: log}
}

// Interceptor creates an audit logging interceptor
func (al *AuditLogger) Interceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Extract metadata
		userID := "anonymous"
		email := ""
		roles := []string{}
		clientIP := ""

		if claims, ok := GetClaims(ctx); ok {
			userID = GetUserID(claims)
			email = GetEmail(claims)
			roles = GetRoles(claims)
		}

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
				clientIP = ips[0]
			} else if ips := md.Get("x-real-ip"); len(ips) > 0 {
				clientIP = ips[0]
			}
		}

		// Call handler
		resp, err := handler(ctx, req)

		// Log the event
		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			}
		}

		// Log with appropriate level based on outcome
		attrs := []any{
			slog.String("event", "grpc_request"),
			slog.String("method", info.FullMethod),
			slog.String("user_id", userID),
			slog.String("email", email),
			slog.Any("roles", roles),
			slog.String("client_ip", clientIP),
			slog.String("status", statusCode.String()),
			slog.Duration("duration", duration),
			slog.Time("timestamp", start),
		}

		switch {
		case statusCode == codes.Unauthenticated:
			al.log.Warn("Authentication failed", attrs...)
		case statusCode == codes.PermissionDenied:
			al.log.Warn("Authorization failed", attrs...)
		case err != nil:
			al.log.Error("Request failed", append(attrs, slog.String("error", err.Error()))...)
		default:
			// Log successful authenticated requests at info level
			if userID != "anonymous" {
				al.log.Info("Authenticated request", attrs...)
			}
		}

		return resp, err
	}
}
