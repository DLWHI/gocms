package tokens

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Username    string `json:"username"`
	Permissions uint32 `json:"perm"`
	jwt.RegisteredClaims
}

type Token string

func GenerateToken(username string, perm uint32, jwtSecret *rsa.PrivateKey, duration time.Duration) (Token, error) {
	claims := JWTClaims{
		Username:    username,
		Permissions: perm,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gocms",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return Token(signed), nil
}
