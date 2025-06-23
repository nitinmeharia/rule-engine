package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	ClientID string `json:"clientId"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	var (
		secret   = flag.String("secret", "change-me-in-production", "JWT secret key")
		clientID = flag.String("client-id", "test-client-123", "Client ID")
		role     = flag.String("role", "admin", "User role (admin, viewer, executor)")
		expiry   = flag.Duration("expiry", 24*time.Hour, "Token expiration time")
	)
	flag.Parse()

	// Create claims
	claims := JWTClaims{
		ClientID: *clientID,
		Role:     *role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(*expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(*secret))
	if err != nil {
		log.Fatalf("Failed to sign token: %v", err)
	}

	fmt.Printf("JWT Token:\n%s\n\n", tokenString)
	fmt.Printf("Token Details:\n")
	fmt.Printf("  Client ID: %s\n", *clientID)
	fmt.Printf("  Role: %s\n", *role)
	fmt.Printf("  Expires: %s\n", time.Now().Add(*expiry).Format(time.RFC3339))
	fmt.Printf("\nUsage:\n")
	fmt.Printf("  curl -H \"Authorization: Bearer %s\" http://localhost:8080/v1/namespaces\n", tokenString)
}
