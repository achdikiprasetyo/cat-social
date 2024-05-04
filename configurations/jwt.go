package configurations

import (
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var JWTSecret string = os.Getenv("DB_HOST")

func GetUserFromToken(c *gin.Context) (int, error) {
	token := c.GetHeader("Authorization")
	fmt.Println("Token received:", token) // Print the received token
	if token == "" {
		return 0, fmt.Errorf("missing bearer token")
	}

	// Remove "Bearer " prefix from the token string
	token = token[7:]

	// Parse the token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil // Change this to your secret key
	})
	if err != nil {
		return 0, fmt.Errorf("invalid bearer token: %v", err)
	}

	// Extract the user ID from claims
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid user ID in token")
	}

	// Convert float64 to int
	userID := int(userIDFloat)

	fmt.Println(userID)

	return userID, nil
}

func GenerateToken(email string, userID int) (string, error) {
	// Create the Claims
	claims := jwt.MapClaims{
		"email":   email,
		"user_id": userID,                               // Add user ID to the claims
		"exp":     time.Now().Add(time.Hour * 8).Unix(), // Token expiration time, adjust as needed
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key
	secretKey := []byte(JWTSecret) // Change this to your secret key
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func CheckBearerToken(c *gin.Context) error {
	token := c.GetHeader("Authorization")
	fmt.Println("Token received:", token) // Print the received token
	if token == "" {
		return fmt.Errorf("missing bearer token")
	}

	// Remove "Bearer " prefix from the token string
	token = token[7:]

	// Parse the token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil // Change this to your secret key
	})
	if err != nil {
		return fmt.Errorf("invalid bearer token: %v", err)
	}

	// Token is valid
	return nil
}
