package middleware

import "crypto/rsa"

func NewJWTApiKey(jwtSecret *rsa.PublicKey) *JWTAuth {
	return &JWTAuth{
		jwtSecret: jwtSecret,
	}
}
