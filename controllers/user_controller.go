package controllers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/dgrijalva/jwt-go"
)

// UserController represents the user controller
type UserController struct {
	DB *sql.DB
}

// NewUserController creates a new user controller
func NewUserController(db *sql.DB) *UserController {
	return &UserController{DB: db}
}

// Register registers a new user
func (uc *UserController) Register(c *gin.Context) {
	var user struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required,min=5,max=50"`
		Password string `json:"password" binding:"required,min=5,max=15"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Save user to database
	_, err = uc.DB.Exec("INSERT INTO users (email, name, password) VALUES ($1, $2, $3)", user.Email, user.Name, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "access_token": token})
}

// Login logs in a user
func (uc *UserController) Login(c *gin.Context) {
	var loginReq struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=5,max=15"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch user from database
	var storedPassword string
	err := uc.DB.QueryRow("SELECT password FROM users WHERE email = $1", loginReq.Email).Scan(&storedPassword)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(loginReq.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password"})
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(loginReq.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User logged successfully", "access_token": token})
}

// GenerateJWT generates a JWT token for the given email
func GenerateJWT(email string) (string, error) {
	// Create the Claims
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 8).Unix(), // Token expiration time, adjust as needed
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key
	secretKey := []byte("your-secret-key") // Change this to your secret key
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
