package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT token claims
type Claims struct {
	ClientID string `json:"clientId"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	secret := "dev-secret-key-change-in-production"

	// Create claims
	claims := Claims{
		ClientID: "test-client",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Fatal("Error signing token:", err)
	}

	fmt.Println("Generated JWT token:")
	fmt.Println(tokenString)
}
