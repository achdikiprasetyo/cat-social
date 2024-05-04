package main

import (
	"CatsSocial/controllers"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// DB represents the database connection
var DB *sql.DB

func NewCatController(db *sql.DB) *CatController {
	return &CatController{DB: db}
}

func main() {
	// Baca nilai variabel lingkungan untuk koneksi database

	// Inisialisasi router Gin
	router := gin.Default()

	// Atur rute untuk register dan login
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // Atur origin sesuai kebutuhan Anda, "*" untuk memperbolehkan dari semua origin
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(config))

	// Atur rute untuk register dan login
	router.POST("/v1/user/register", controllers.Register)
	router.POST("/v1/user/login", controllers.Login)

	router.POST("/v1/cat", controllers.CreateCat)
	router.GET("/v1/cat", controllers.GetCats)
	router.PUT("/v1/cat/:id", controllers.UpdateCat)
	router.DELETE("/v1/cat/:id", controllers.DeleteCat)

	router.POST("/v1/cat/match", controllers.CreateMatch)
	router.GET("/v1/cat/match", controllers.GetMatchRequests)
	router.POST("/v1/cat/match/approve", controllers.ApproveMatch)
	router.POST("/v1/cat/match/reject", controllers.RejectMatch)
	router.DELETE("/v1/cat/match/:id", controllers.DeleteMatch)
	// Jalankan server HTTP
	router.Run(":8080")
}

type CatController struct {
	DB *sql.DB
}

func GetUserIDFromToken(c *gin.Context) (int, error) {
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
		return []byte("9#JKl!M8Pn$1Sd@5"), nil // Change this to your secret key
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

// CheckBearerToken checks the bearer token
// CheckBearerToken checks the bearer token
func CheckBearerToken(c *gin.Context) error {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return fmt.Errorf("missing bearer token")
	}

	// Split token string
	parts := strings.Split(tokenString, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return fmt.Errorf("invalid bearer token format")
	}
	token := parts[1]

	// Parse the token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("9#JKl!M8Pn$1Sd@5"), nil // Change this to your secret key
	})
	if err != nil {
		return fmt.Errorf("invalid bearer token")
	}

	// Token is valid
	return nil
}

// UserController represents the user controller
type UserController struct {
	DB *sql.DB
}

// GenerateJWT generates a JWT token for the given email
// GenerateJWT generates a JWT token for the given email and user ID
func GenerateJWT(email string, userID int) (string, error) {
	// Create the Claims
	claims := jwt.MapClaims{
		"email":   email,
		"user_id": userID,                               // Add user ID to the claims
		"exp":     time.Now().Add(time.Hour * 8).Unix(), // Token expiration time, adjust as needed
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

func (cc *CatController) CreateMatch(c *gin.Context) {
	// Mendapatkan user ID dari token JWT
	userID, err := GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind request body
	var matchRequest struct {
		MatchCatID string `json:"matchCatId" binding:"required"`
		UserCatID  string `json:"userCatId" binding:"required"`
		Message    string `json:"message" binding:"required,min=5,max=120"`
	}

	if err := c.ShouldBindJSON(&matchRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah cat yang dimaksud milik pengguna
	var ownerID int
	err = cc.DB.QueryRow("SELECT user_id FROM cats WHERE id = $1", matchRequest.UserCatID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User cat not found"})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Cek apakah gender kucing sama
	var userCatSex string
	var matchCatSex string
	err = cc.DB.QueryRow("SELECT sex FROM cats WHERE id = $1", matchRequest.UserCatID).Scan(&userCatSex)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User cat not found"})
		return
	}
	err = cc.DB.QueryRow("SELECT sex FROM cats WHERE id = $1", matchRequest.MatchCatID).Scan(&matchCatSex)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match cat not found"})
		return
	}

	if userCatSex == matchCatSex {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both cats have the same gender"})
		return
	}

	// Cek apakah kedua kucing sudah dipasangkan sebelumnya
	var isMatched bool
	err = cc.DB.QueryRow("SELECT status FROM match_cats WHERE (issuedCatId = $1 AND receiverCatId = $2) OR (issuedCatId = $2 AND receiverCatId = $1)", matchRequest.UserCatID, matchRequest.MatchCatID).Scan(&isMatched)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check matching status"})
		return
	}
	if isMatched {
		c.JSON(http.StatusBadRequest, gin.H{"error": "One of the cats has already been matched"})
		return
	}

	// Cek apakah kedua kucing milik pemilik yang sama
	var userCatOwnerID int
	var matchCatOwnerID int
	err = cc.DB.QueryRow("SELECT user_id FROM cats WHERE id = $1", matchRequest.UserCatID).Scan(&userCatOwnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User cat not found"})
		return
	}
	err = cc.DB.QueryRow("SELECT user_id FROM cats WHERE id = $1", matchRequest.MatchCatID).Scan(&matchCatOwnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match cat not found"})
		return
	}
	if userCatOwnerID == matchCatOwnerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both cats belong to the same owner"})
		return
	}

	// Tambahkan permintaan pencocokan kucing ke database
	_, err = cc.DB.Exec("INSERT INTO match_cats (issuedId, issuedCatId, receiverId, receiverCatId, message) VALUES ($1, $2, $3, $4, $5)",
		userID, matchRequest.UserCatID, matchCatOwnerID, matchRequest.MatchCatID, matchRequest.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add match request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Match request sent successfully"})
}

type User struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

// CatDetail represents the cat details structure
type CatDetail struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Race        string   `json:"race"`
	Sex         string   `json:"sex"`
	AgeInMonth  int      `json:"ageInMonth"`
	Description string   `json:"description"`
	ImageURLs   []string `json:"imageUrls"`
}

// CatController dan fungsi lainnya tetap sama seperti sebelumnya

// GetMatchRequests retrieves match requests from the database
func (cc *CatController) GetMatchRequests(c *gin.Context) {
	// Mendapatkan user ID dari token JWT
	userID, err := GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Retrieve match requests from the database
	rows, err := cc.DB.Query("SELECT mc.id, u1.name AS issuedByName, u1.email AS issuedByEmail, u1.created_at AS issuedAt, u2.name AS receiverName, u2.email AS receiverEmail, u2.created_at AS receivedAt, c1.id AS issuedCatId, c1.name AS issuedCatName, c1.race AS issuedCatRace, c1.sex AS issuedCatSex, c1.age_in_month AS issuedCatAgeInMonth, c1.description AS issuedCatDescription, c1.image_urls AS issuedCatImageUrls, c2.id AS receiverCatId, c2.name AS receiverCatName, c2.race AS receiverCatRace, c2.sex AS receiverCatSex, c2.age_in_month AS receiverCatAgeInMonth, c2.description AS receiverCatDescription, c2.image_urls AS receiverCatImageUrls, mc.message, mc.created_at FROM match_cats mc INNER JOIN users u1 ON mc.issuedId = u1.id INNER JOIN users u2 ON mc.receiverId = u2.id INNER JOIN cats c1 ON mc.issuedCatId = c1.id INNER JOIN cats c2 ON mc.receiverCatId = c2.id WHERE mc.issuedId = $1 OR mc.receiverId = $1 ORDER BY mc.created_at DESC", userID)
	if err != nil {
		log.Println("Error retrieving match requests:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match requests"})
		return
	}
	defer rows.Close()

	// Create slice to hold retrieved match requests
	var matchRequests []gin.H

	// Iterate over rows and append to matchRequests slice
	for rows.Next() {
		var matchRequest struct {
			ID             int       `json:"id"`
			IssuedBy       User      `json:"issuedBy"`
			MatchCatDetail CatDetail `json:"matchCatDetail"`
			UserCatDetail  CatDetail `json:"userCatDetail"`
			Message        string    `json:"message"`
			CreatedAt      time.Time `json:"createdAt"`
		}
		err := rows.Scan(
			&matchRequest.ID,
			&matchRequest.IssuedBy.Name,
			&matchRequest.IssuedBy.Email,
			&matchRequest.IssuedBy.CreatedAt,
			&matchRequest.UserCatDetail.ID, // Ambil ID dari UserCatDetail
			&matchRequest.UserCatDetail.Name,
			&matchRequest.UserCatDetail.Race,
			&matchRequest.UserCatDetail.Sex,
			&matchRequest.UserCatDetail.AgeInMonth,
			&matchRequest.UserCatDetail.Description,
			&matchRequest.UserCatDetail.ImageURLs,
			&matchRequest.MatchCatDetail.ID, // Ambil ID dari MatchCatDetail
			&matchRequest.MatchCatDetail.Name,
			&matchRequest.MatchCatDetail.Race,
			&matchRequest.MatchCatDetail.Sex,
			&matchRequest.MatchCatDetail.AgeInMonth,
			&matchRequest.MatchCatDetail.Description,
			&matchRequest.MatchCatDetail.ImageURLs,
			&matchRequest.Message,
			&matchRequest.CreatedAt,
		)

		if err != nil {
			log.Println("Error scanning row:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match requests"})
			return
		}

		matchRequests = append(matchRequests, gin.H{
			"id":             matchRequest.ID,
			"issuedBy":       matchRequest.IssuedBy,
			"matchCatDetail": matchRequest.MatchCatDetail,
			"userCatDetail":  matchRequest.UserCatDetail,
			"message":        matchRequest.Message,
			"createdAt":      matchRequest.CreatedAt.Format(time.RFC3339),
		})
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "data": matchRequests})
}

func (cc *CatController) ApproveMatch(c *gin.Context) {
	// Get user ID from JWT token
	userID, err := GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind request body
	var approval struct {
		MatchID string `json:"matchId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&approval); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if match request exists in the database and the user is the issuer
	var matchCount int
	err = cc.DB.QueryRow("SELECT COUNT(*) FROM match_cats WHERE id = $1 AND issuedId = $2", approval.MatchID, userID).Scan(&matchCount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	}

	if matchCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	}

	// Approve cat matching request
	_, err = cc.DB.Exec("UPDATE match_cats SET status = true WHERE id = $1", approval.MatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve match request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match request approved successfully"})
}

func (cc *CatController) RejectMatch(c *gin.Context) {
	// Get user ID from JWT token
	userID, err := GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind request body
	var rejection struct {
		MatchID string `json:"matchId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&rejection); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if match request exists in the database and the user is the issuer
	var matchCount int
	err = cc.DB.QueryRow("SELECT COUNT(*) FROM match_cats WHERE id = $1 AND issuedId = $2", rejection.MatchID, userID).Scan(&matchCount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	}

	if matchCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	}

	// Reject cat matching request
	_, err = cc.DB.Exec("DELETE FROM match_cats WHERE id = $1", rejection.MatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject match request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match request rejected successfully"})
}
func (cc *CatController) DeleteMatch(c *gin.Context) {
	// Mendapatkan user ID dari token JWT
	userID, err := GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Mendapatkan ID pencocokan dari path params
	matchID := c.Param("id")

	// Cek apakah pemintaan pencocokan ada dalam database
	var matchCount int
	err = cc.DB.QueryRow("SELECT COUNT(*) FROM match_cats WHERE id = $1 AND issuedId = $2", matchID, userID).Scan(&matchCount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	}

	if matchCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	}

	// Hapus permintaan pencocokan kucing
	_, err = cc.DB.Exec("DELETE FROM match_cats WHERE id = $1", matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete match request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match request deleted successfully"})
}
