package controllers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	
)

// CatController represents the cat controller
type CatController struct {
	DB *sql.DB
}

// NewCatController creates a new cat controller
func NewCatController(db *sql.DB) *CatController {
	return &CatController{DB: db}
}

// CreateCat creates a new cat in the database
func (cc *CatController) CreateCat(c *gin.Context) {
	// Check bearer token
	if err := CheckBearerToken(c); err != nil {
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

	// Insert cat data into database
	_, err := cc.DB.Exec("INSERT INTO cats (name, race, sex, age_in_month, description, image_urls) VALUES ($1, $2, $3, $4, $5, $6)",
		cat.Name, cat.Race, cat.Sex, cat.AgeInMonth, cat.Description, cat.ImageURLs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add cat"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Cat added successfully"})
}

// CheckBearerToken checks the bearer token
func CheckBearerToken(c *gin.Context) error {
	token := c.GetHeader("Authorization")
	if token == "" {
		return errors.New("missing bearer token")
	}

	// Perform token validation here
	// You need to implement token validation logic according to your application requirements

	return nil
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

	// Update cat data in the database
	_, err := cc.DB.Exec("UPDATE cats SET name=$1, race=$2, sex=$3, age_in_month=$4, description=$5, image_urls=$6 WHERE id=$7",
		cat.Name, cat.Race, cat.Sex, cat.AgeInMonth, cat.Description, cat.ImageURLs, catID)
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
