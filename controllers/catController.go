package controllers

import (
	"CatsSocial/configurations"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func CreateCat(c *gin.Context) {
	// Mendapatkan user ID dari token JWT
	userID, err := configurations.GetUserFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
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

	var catID int
	var createdAt time.Time
	err = DB.QueryRow("INSERT INTO cats (name, race, sex, age_in_month, description, image_urls, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at",
		cat.Name, cat.Race, cat.Sex, cat.AgeInMonth, cat.Description, pq.Array(cat.ImageURLs), userID).Scan(&catID, &createdAt)
	if err != nil {
		log.Println("Error adding cat:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add cat"})
		return
	}

	// Construct the JSON response
	response := gin.H{
		"message": "Success",
		"data": gin.H{
			"id":        catID,
			"createdAt": createdAt.Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusCreated, response)
}

func GetCats(c *gin.Context) {
	if err := configurations.CheckBearerToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Retrieve cats from the database
	rows, err := DB.Query("SELECT id, name, race, sex, age_in_month, description, image_urls, has_matched, created_at FROM cats ORDER BY created_at DESC")
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
			ID          int       `json:"id"`
			Name        string    `json:"name"`
			Race        string    `json:"race"`
			Sex         string    `json:"sex"`
			AgeInMonth  int       `json:"ageInMonth"`
			Description string    `json:"description"`
			ImageURLs   []string  `json:"imageUrls"`
			HasMatched  bool      `json:"hasMatched"`
			CreatedAt   time.Time `json:"createdAt"`
		}

		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Race, &cat.Sex, &cat.AgeInMonth, &cat.Description, pq.Array(&cat.ImageURLs), &cat.HasMatched, &cat.CreatedAt); err != nil {
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
			"imageUrls":   cat.ImageURLs,
			"description": cat.Description,
			"hasMatched":  cat.HasMatched,
			"createdAt":   cat.CreatedAt.Format(time.RFC3339),
		})
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cats"})
		return
	}

	// Construct the response JSON
	response := gin.H{
		"message": "success",
		"data":    cats,
	}

	c.JSON(http.StatusOK, response)
}

func UpdateCat(c *gin.Context) {
	userID, err := configurations.GetUserFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get cat ID from path params
	catID := c.Param("id")
	var catUserId int
	err = DB.QueryRow("SELECT user_id from cats where id=$1", catID).Scan(&catUserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if catUserId != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not Allowed to Update"})
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

	// Update cat data in the database
	_, err = DB.Exec("UPDATE cats SET name=$1, race=$2, sex=$3, age_in_month=$4, description=$5, image_urls=$6 WHERE id=$7",
		cat.Name, cat.Race, cat.Sex, cat.AgeInMonth, cat.Description, pq.Array(cat.ImageURLs), catID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cat updated successfully"})
}

func DeleteCat(c *gin.Context) {
	userID, err := configurations.GetUserFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	catID := c.Param("id")

	var catUserID int
	err = DB.QueryRow("SELECT user_id FROM cats WHERE id=$1", catID).Scan(&catUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cat not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if catUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to delete"})
		return
	}

	// Delete cat from the database
	_, err = DB.Exec("DELETE FROM cats WHERE id=$1", catID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cat deleted successfully"})
}
