package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrTokenExpired         = errors.New("token expired")
)

type TokenManager struct {
	PublicKey *rsa.PublicKey
}

type Claims struct {
	UserID    int64  `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func NewTokenManagerRSA(publicKeyPath string) (*TokenManager, error) {
	pubData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key from %s: %v", publicKeyPath, err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubData)
	if err != nil {
		return nil, err
	}

	return &TokenManager{
		PublicKey: pubKey,
	}, nil

}

func (tm *TokenManager) ParseToken(tokenStr string) (*Claims, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, ErrInvalidSigningMethod
		}
		return tm.PublicKey, nil
	}
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, keyFunc)

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.ExpiresAt.Before(time.Now()) {
			return nil, ErrTokenExpired
		}
		return claims, nil
	}
	return nil, fmt.Errorf("parse token: %w", err)
}
