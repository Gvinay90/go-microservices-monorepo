package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrInsufficientPerms = errors.New("insufficient permissions")
	ErrMissingToken      = errors.New("missing authorization token")
)

// Claims represents JWT claims
type Claims struct {
	jwt.RegisteredClaims
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

// JWKSKey represents a JSON Web Key
type JWKSKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents JSON Web Key Set
type JWKS struct {
	Keys []JWKSKey `json:"keys"`
}

// Validator validates JWT tokens
type Validator struct {
	keycloakURL string
	realm       string
	keys        map[string]*rsa.PublicKey
	keysMux     sync.RWMutex
	httpClient  *http.Client
}

// NewValidator creates a new JWT validator
func NewValidator(keycloakURL, realm string) *Validator {
	return &Validator{
		keycloakURL: keycloakURL,
		realm:       realm,
		keys:        make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateToken validates a JWT token and returns claims
func (v *Validator) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	// Parse token without verification first to get the kid
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}

		// Get public key
		publicKey, err := v.getPublicKey(ctx, kid)
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// getPublicKey retrieves the public key for the given kid
func (v *Validator) getPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	// Check cache first
	v.keysMux.RLock()
	if key, exists := v.keys[kid]; exists {
		v.keysMux.RUnlock()
		return key, nil
	}
	v.keysMux.RUnlock()

	// Fetch JWKS
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", v.keycloakURL, v.realm)

	req, err := http.NewRequestWithContext(ctx, "GET", jwksURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Find the key with matching kid
	for _, key := range jwks.Keys {
		if key.Kid == kid {
			publicKey, err := parseRSAPublicKey(key)
			if err != nil {
				return nil, err
			}

			// Cache the key
			v.keysMux.Lock()
			v.keys[kid] = publicKey
			v.keysMux.Unlock()

			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// parseRSAPublicKey converts JWKS key to RSA public key
func parseRSAPublicKey(key JWKSKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes).Int64()

	return &rsa.PublicKey{
		N: n,
		E: int(e),
	}, nil
}

// HasRole checks if claims contain a specific role
func HasRole(claims *Claims, role string) bool {
	for _, r := range claims.RealmAccess.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// GetRoles extracts roles from claims
func GetRoles(claims *Claims) []string {
	return claims.RealmAccess.Roles
}

// GetUserID extracts user ID from claims (using subject)
func GetUserID(claims *Claims) string {
	return claims.Subject
}

// GetEmail extracts email from claims
func GetEmail(claims *Claims) string {
	return claims.Email
}

// ExtractToken extracts Bearer token from Authorization header
func ExtractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrMissingToken
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}
