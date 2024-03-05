package jwt

import (
	"errors"
	"time"

	JWT "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	JWT.RegisteredClaims
	Value string `json:"value"`
}

func ParseToken(tokenstr, secret string) (string, error) {
	t, err := JWT.ParseWithClaims(tokenstr, &Claims{}, func(t *JWT.Token) (any, error) {
		if _, ok := t.Method.(*JWT.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := t.Claims.(*Claims)
	if !ok {
		return "", errors.New("token claims are not of type")
	}

	return claims.Value, nil
}

func GenerateToken(value, secret string, tokenTTL time.Duration) (string, error) {
	registeredClaims := JWT.RegisteredClaims{
		ExpiresAt: &JWT.NumericDate{Time: time.Now().Add(tokenTTL)},
		IssuedAt:  &JWT.NumericDate{Time: time.Now()}}

	token := JWT.NewWithClaims(JWT.SigningMethodHS256, &Claims{registeredClaims, value})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.New("can't create token")
	}
	return tokenString, nil
}
