bukan itu buatkan kodenya seusi kontrak itu, dan tuliskan perubahannya ke kode saya package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/lib/pq" 
	"golang.org/x/crypto/bcrypt"
	"github.com/dgrijalva/jwt-go"
)

// DB represents the database connection
var DB *sql.DB

func NewCatController(db *sql.DB) *CatController {
	return &CatController{DB: db}
}

func main() {
	// Baca nilai variabel lingkungan untuk koneksi database
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Buat string koneksi database
	dbConnectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Buat koneksi ke database
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Tes koneksi ke database
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("Connected to database")

	// Simpan koneksi database untuk digunakan di seluruh aplikasi
	DB = db

	// Inisialisasi router Gin
	router := gin.Default()

	// Atur rute untuk register dan login
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // Atur origin sesuai kebutuhan Anda, "*" untuk memperbolehkan dari semua origin
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(config))

	// Atur rute untuk register dan login
	userController := NewUserController(DB)
	router.POST("/v1/user/register", userController.Register)
	router.POST("/v1/user/login", userController.Login)

	catController := NewCatController(DB)
	router.POST("/v1/cat", catController.CreateCat)
	router.GET("/v1/cat", catController.GetCats)
	router.PUT("/v1/cat/:id", catController.UpdateCat)
	router.DELETE("/v1/cat/:id", catController.DeleteCat)

	// Jalankan server HTTP
	router.Run(":32")
}

type CatController struct {
	DB *sql.DB
}

// CreateCat creates a new cat in the database
// CreateCat creates a new cat in the database
func (cc *CatController) CreateCat(c *gin.Context) {
    // Mendapatkan user ID dari token JWT
    userID, err := GetUserIDFromToken(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Bind request body to Cat struct
    var cat struct {
        Name        string   `json:"name" binding:"required,min=1,max=30"`
        Race        string   `json:"race" binding:"required,oneof=Persian MaineCoon Siamese Ragdoll Bengal Sphynx BritishShorthair Abyssinian ScottishFold Birman"`
        Sex         string   `json:"sex" binding:"required,oneof=male female"`
        AgeInMonth  int      `json:"ageInMonth" binding:"required,min=1,max=120082"`
        Description string   `json:"description" binding:"required,min=1,max=200"`
        ImageURLs   []string `json:"imageUrls" binding:"required,min=1,dive,url"`
    }

    if err := c.ShouldBindJSON(&cat); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    _, err = cc.DB.Exec("INSERT INTO cats (name, race, sex, age_in_month, description, image_urls, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7)",
        cat.Name, cat.Race, cat.Sex, cat.AgeInMonth, cat.Description, pq.Array(cat.ImageURLs), userID)
    if err != nil {
        log.Println("Error adding cat:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add cat"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Cat added successfully"})
}


func GetUserIDFromToken(c *gin.Context) (int, error) {
    token := c.GetHeader("Authorization")
    if token == "" {
        return 0, fmt.Errorf("missing bearer token")
    }

    // Parse the token
    claims := jwt.MapClaims{}
    _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte("9#JKl!M8Pn$1Sd@5"), nil // Change this to your secret key
    })
    if err != nil {
        return 0, fmt.Errorf("invalid bearer token")
    }

    // Extract the user ID from claims
    userIDFloat, ok := claims["user_id"].(float64)
if !ok {
    return 0, fmt.Errorf("invalid user ID in token")
}

// Konversi float64 ke int
userID := int(userIDFloat)

    if !ok {
        return 0, fmt.Errorf("invalid user ID in token")
    }

    return userID, nil
}	


// UpdateCat updates an existing cat in the database
func (cc *CatController) UpdateCat(c *gin.Context) {
	// Check bearer token
	if err := CheckBearerToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get cat ID from path params
	catID := c.Param("id")

	// Bind request body to Cat struct
	var cat struct {
		Name        string   `json:"name" binding:"required,min=1,max=30"`
		Race        string   `json:"race" binding:"required,oneof=Persian MaineCoon Siamese Ragdoll Bengal Sphynx BritishShorthair Abyssinian ScottishFold Birman"`
		Sex         string   `json:"sex" binding:"required,oneof=male female"`
		AgeInMonth  int      `json:"ageInMonth" binding:"required,min=1,max=120082"`
		Description string   `json:"description" binding:"required,min=1,max=200"`
		ImageURLs   []string `json:"imageUrls" binding:"required,min=1,dive,url"`
	}

	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mengonversi array []string menjadi string JSON
	imageURLsJSON, err := json.Marshal(cat.ImageURLs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal image URLs"})
		return
	}

	// Update cat data in the database
	_, err = cc.DB.Exec("UPDATE cats SET name=$1, race=$2, sex=$3, age_in_month=$4, description=$5, image_urls=$6 WHERE id=$7",
		cat.Name, cat.Race, cat.Sex, cat.AgeInMonth, cat.Description, string(imageURLsJSON), catID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cat updated successfully"})
}

// DeleteCat deletes an existing cat from the database
func (cc *CatController) DeleteCat(c *gin.Context) {
	// Check bearer token
	if err := CheckBearerToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get cat ID from path params
	catID := c.Param("id")

	// Delete cat from the database
	_, err := cc.DB.Exec("DELETE FROM cats WHERE id=$1", catID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cat deleted successfully"})
}

// GetCats retrieves all cats from the database
func (cc *CatController) GetCats(c *gin.Context) {
    // Check bearer token
    if err := CheckBearerToken(c); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Retrieve cats from the database
    rows, err := cc.DB.Query("SELECT id, name, race, sex, age_in_month, description, image_urls FROM cats")
    if err != nil {
        log.Println("Error retrieving cats:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cats"})
        return
    }
    defer rows.Close()

    // Create slice to hold retrieved cats
    var cats []gin.H

    // Iterate over rows and append to cats slice
    for rows.Next() {
        var cat struct {
            ID          int      `json:"id"`
            Name        string   `json:"name"`
            Race        string   `json:"race"`
            Sex         string   `json:"sex"`
            AgeInMonth  int      `json:"ageInMonth"`
            Description string   `json:"description"`
            ImageURLs   []string `json:"imageUrls"`
        }

        if err := rows.Scan(&cat.ID, &cat.Name, &cat.Race, &cat.Sex, &cat.AgeInMonth, &cat.Description, pq.Array(&cat.ImageURLs)); err != nil {
            log.Println("Error scanning row:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cats"})
            return
        }

        cats = append(cats, gin.H{
            "id":          cat.ID,
            "name":        cat.Name,
            "race":        cat.Race,
            "sex":         cat.Sex,
            "ageInMonth":  cat.AgeInMonth,
            "description": cat.Description,
            "imageUrls":   cat.ImageURLs,
        })
    }

    // Check for errors during rows iteration
    if err := rows.Err(); err != nil {
        log.Println("Error iterating over rows:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cats"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"cats": cats})
}

// CheckBearerToken checks the bearer token
func CheckBearerToken(c *gin.Context) error {
	token := c.GetHeader("Authorization")
	if token == "" {
		return fmt.Errorf("missing bearer token")
	}

	// Perform token validation here
	// You need to implement token validation logic according to your application requirements

	return nil
}

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

    // Retrieve user ID from the database (assuming the ID is auto-incremented)
    var userID int
    err = uc.DB.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID"})
        return
    }

    // Generate JWT token with user ID
    token, err := GenerateJWT(user.Email, userID)
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
    var userID int // Define variable to store user ID
    err := uc.DB.QueryRow("SELECT id, password FROM users WHERE email = $1", loginReq.Email).Scan(&userID, &storedPassword)
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
    token, err := GenerateJWT(loginReq.Email, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User logged successfully", "access_token": token})
}

// GenerateJWT generates a JWT token for the given email
// GenerateJWT generates a JWT token for the given email and user ID
func GenerateJWT(email string, userID int) (string, error) {
    // Create the Claims
    claims := jwt.MapClaims{
        "email":  email,
        "user_id": userID, // Add user ID to the claims
        "exp":    time.Now().Add(time.Hour * 8).Unix(), // Token expiration time, adjust as needed
    }

    // Create the token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Sign the token with a secret key
    secretKey := []byte("9#JKl!M8Pn$1Sd@5") // Change this to your secret key
    tokenString, err := token.SignedString(secretKey)
    if err != nil {
        return "", err
    }

    return tokenString, nil
}


