package helpers

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/config"
)

type (
	jwtHelper struct {
		secretKey string
		duration  int
	}

	JwtHelper interface {
		GenerateToken(payload map[string]interface{}) (string, error)
		ValidateToken(token string) (map[string]interface{}, error)
	}
)

type JWTHelperParams struct {
	fx.In

	Config config.JWTConfig
}

var _ JwtHelper = (*jwtHelper)(nil)

func NewJWTHelper(params JWTHelperParams) JwtHelper {
	return &jwtHelper{
		secretKey: params.Config.SecretKey,
		duration:  params.Config.Duration,
	}
}

func (h *jwtHelper) GenerateToken(payload map[string]interface{}) (string, error) {
	now := time.Now()

	baseClaims := jwt.MapClaims{
		"iss": "federation",                                            // Issuer
		"iat": now.Unix(),                                              // Issued At
		"exp": now.Add(time.Duration(h.duration) * time.Second).Unix(), // Expired At
	}

	// Merge payload with base claims
	// Prevent overwriting of protected claims
	protectedClaims := map[string]bool{"iss": true, "iat": true}
	for key, value := range payload {
		if contains(protectedClaims, key) {
			continue
		}
		baseClaims[key] = value
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, baseClaims)

	return claims.SignedString([]byte(h.secretKey))
}

func (h *jwtHelper) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// Helper function to check if a string exists in slice
func contains(m map[string]bool, item string) bool {
	if v, ok := m[item]; ok && v {
		return true
	}
	return false
}
