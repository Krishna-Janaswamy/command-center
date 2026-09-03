package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtService struct {
	secret string
}

type CustomClaims struct {
	Role    string `json:"role"`
	AdGroup string `json:"adGroup"`
	jwt.RegisteredClaims
}

func NewJwtService(secret string) *JwtService {
	return &JwtService{secret: secret}
}

func (j *JwtService) GenerateToken(username, role, adGroup string) (string, error) {
	claims := CustomClaims{
		Role:    role,
		AdGroup: adGroup,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JwtService) ValidateToken(tokenStr string) (*CustomClaims, error) {
	if tokenStr == "" {
		return nil, errors.New("empty token")
	}
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
