package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"devhive-backend/config"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims matches exact .NET JWT structure
// Must maintain: sub, email, jti, nameidentifier (user.ID)
type JWTClaims struct {
	Sub            string `json:"sub"`            // username
	Email          string `json:"email"`          // user email
	JTI            string `json:"jti"`            // token ID
	NameIdentifier string `json:"nameidentifier"` // user.ID (matches .NET ClaimTypes.NameIdentifier)
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration matching .NET settings
type JWTConfig struct {
	Secret     string
	Issuer     string
	Audience   string
	ExpiryTime time.Duration // 24 hours to match .NET
}

// NewJWTConfig creates JWT config from environment
func NewJWTConfig() *JWTConfig {
	return &JWTConfig{
		Secret:     config.GetEnv("JWT_SECRET", "your-secret-key"),
		Issuer:     config.GetEnv("JWT_ISSUER", "devhive-backend"),
		Audience:   config.GetEnv("JWT_AUDIENCE", "devhive-frontend"),
		ExpiryTime: 24 * time.Hour, // Exact match to .NET 24h expiry
	}
}

// JWTService handles JWT operations with .NET compatibility
type JWTService struct {
	config *JWTConfig
}

// NewJWTService creates new JWT service
func NewJWTService(config *JWTConfig) *JWTService {
	return &JWTService{
		config: config,
	}
}

// GenerateToken creates JWT token with exact .NET claims structure
func (j *JWTService) GenerateToken(userID, username, email string) (string, error) {
	now := time.Now()

	claims := &JWTClaims{
		Sub:            username,      // .NET: ClaimTypes.Name
		Email:          email,         // .NET: ClaimTypes.Email
		JTI:            generateJTI(), // .NET: Jti claim
		NameIdentifier: userID,        // .NET: ClaimTypes.NameIdentifier (user.ID)
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Audience:  []string{j.config.Audience},
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.ExpiryTime)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

// ValidateToken validates JWT token and returns claims
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method matches .NET HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Validate issuer and audience match .NET expectations
		if claims.Issuer != j.config.Issuer {
			return nil, jwt.ErrTokenMalformed
		}
		if !contains(claims.Audience, j.config.Audience) {
			return nil, jwt.ErrTokenMalformed
		}
		return claims, nil
	}

	return nil, jwt.ErrTokenMalformed
}

// AuthMiddleware creates middleware for JWT authentication
// Maintains exact .NET behavior: Authorization: Bearer <token>
func (j *JWTService) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				j.writeUnauthorized(w, "no_authorization_header")
				return
			}

			// Extract token from "Bearer <token>" format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				j.writeUnauthorized(w, "invalid_authorization_format")
				return
			}

			tokenString := parts[1]
			claims, err := j.ValidateToken(tokenString)
			if err != nil {
				j.writeUnauthorized(w, "invalid_token")
				return
			}

			// Add claims to request context for handlers to use
			ctx := context.WithValue(r.Context(), "user_claims", claims)
			ctx = context.WithValue(ctx, "user_id", claims.NameIdentifier)
			ctx = context.WithValue(ctx, "username", claims.Sub)
			ctx = context.WithValue(ctx, "email", claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeUnauthorized writes unauthorized response matching .NET format
func (j *JWTService) writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Helper functions
func generateJTI() string {
	// Generate unique token ID (simplified for now)
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Context helpers for handlers
func GetUserClaims(r *http.Request) *JWTClaims {
	if claims, ok := r.Context().Value("user_claims").(*JWTClaims); ok {
		return claims
	}
	return nil
}

func GetUserID(r *http.Request) string {
	if userID, ok := r.Context().Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func GetUsername(r *http.Request) string {
	if username, ok := r.Context().Value("username").(string); ok {
		return username
	}
	return ""
}

func GetEmail(r *http.Request) string {
	if email, ok := r.Context().Value("email").(string); ok {
		return email
	}
	return ""
}
