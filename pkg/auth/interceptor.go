package auth

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	ClaimsContextKey contextKey = "claims"
)

// Interceptor creates a gRPC authentication interceptor
func Interceptor(validator *Validator, log *slog.Logger, publicMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if method is public (doesn't require authentication)
		if publicMethods != nil && publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Extract token
		token, err := ExtractToken(authHeaders[0])
		if err != nil {
			log.Warn("Failed to extract token", "error", err)
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}

		// Validate token
		claims, err := validator.ValidateToken(ctx, token)
		if err != nil {
			log.Warn("Token validation failed", "error", err, "method", info.FullMethod)
			if err == ErrTokenExpired {
				return nil, status.Error(codes.Unauthenticated, "token expired")
			}
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Inject claims into context
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)

		log.Debug("Request authenticated",
			"method", info.FullMethod,
			"user", claims.PreferredUsername,
			"roles", claims.RealmAccess.Roles,
		)

		return handler(ctx, req)
	}
}

// GetClaims retrieves claims from context
func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}

// RequireRole checks if the user has a specific role
func RequireRole(ctx context.Context, role string) error {
	claims, ok := GetClaims(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "no authentication claims found")
	}

	if !HasRole(claims, role) {
		return status.Errorf(codes.PermissionDenied, "requires role: %s", role)
	}

	return nil
}

// RequireOwnership checks if the user owns the resource
func RequireOwnership(ctx context.Context, resourceUserID string) error {
	claims, ok := GetClaims(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "no authentication claims found")
	}

	// Admins can access anything
	if HasRole(claims, "admin") {
		return nil
	}

	// Check ownership
	if GetUserID(claims) != resourceUserID && GetEmail(claims) != resourceUserID {
		return status.Error(codes.PermissionDenied, "access denied: not resource owner")
	}

	return nil
}
