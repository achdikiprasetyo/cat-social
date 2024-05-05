package controllers

import (
	"CatsSocial/configurations"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	var user struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required,min=5,max=50"`
		Password string `json:"password" binding:"required,min=5,max=15"`
	}

	// c.Bind(&user)
	// c.JSON(http.StatusBadRequest, gin.H{"tes": user.Email})
	// 	return
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User min 5 and max 50 character, Pssword min 5 and max 15 character"})
		return
	}

	rows, err := DB.Query("SELECT * FROM users WHERE email = $1", user.Email)
	if rows.Next() {
		c.JSON(http.StatusConflict, gin.H{"error": "Email Has been used"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// // Save user to database
	_, err = DB.Exec("INSERT INTO users (email, name, password) VALUES ($1, $2, $3)", user.Email, user.Name, string(hashedPassword))
	if err != nil {
		fmt.Println("tes", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Retrieve user ID from the database (assuming the ID is auto-incremented)
	var userID int
	err = DB.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID"})
		return
	}

	// Generate JWT token with user ID
	token, err := configurations.GenerateToken(user.Email, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"data": gin.H{
			"email":       user.Email,
			"name":        user.Name,
			"accessToken": token,
		},
	})
	defer DB.Close()
}

// Login logs in a user
func Login(c *gin.Context) {
	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
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
	var userID int
	var userEmail string
	var userName string
	err = DB.QueryRow(`SELECT id, email, name, password FROM users WHERE email = $1`, loginReq.Email).Scan(&userID, &userEmail, &userName, &storedPassword)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(loginReq.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password"})
		return
	}

	// Generate JWT token with user ID
	token, err := configurations.GenerateToken(loginReq.Email, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User logged successfully",
		"data": gin.H{
			"email":       userEmail,
			"name":        userName,
			"accessToken": token,
		},
	})
	defer DB.Close()
}
