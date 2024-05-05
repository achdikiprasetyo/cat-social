package controllers

import (
	"CatsSocial/configurations"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	// Validasi input
	var cat struct {
		Name        string   `json:"name" binding:"required,min=1,max=30"`
		Race        string   `json:"race" binding:"required,oneof=Persian 'Maine Coon' Siamese Ragdoll Bengal Sphynx 'British Shorthair' 'Abyssinian' 'Scottish Fold' Birman"`
		Sex         string   `json:"sex" binding:"required,oneof='male' 'female'"`
		AgeInMonth  int      `json:"ageInMonth" binding:"required,min=1,max=120082"`
		Description string   `json:"description" binding:"required,min=1,max=200"`
		ImageURLs   []string `json:"imageUrls" binding:"required,min=1,dive,url"`
	}

	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi usia kucing
	if cat.AgeInMonth < 1 || cat.AgeInMonth > 120082 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ageInMonth, must be between 1 and 120082"})
		return
	}

	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
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
	defer DB.Close()

	// Construct the JSON response
	response := gin.H{
		"message": "success",
		"data": gin.H{
			"id":        strconv.Itoa(catID),
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

	// Get query parameters
	id := c.Query("id")
	race := c.Query("race")
	sex := c.Query("sex")
	hasMatchedStr := c.Query("hasMatched")
	ageInMonth := c.Query("ageInMonth")
	ownedStr := c.Query("owned")
	search := c.Query("search")
	limit := c.DefaultQuery("limit", "5")
	offset := c.DefaultQuery("offset", "0")
	var deletedAt string

	// Construct SQL query based on query parameters
	query := "SELECT id, name, race, sex, age_in_month, description, image_urls, has_matched, created_at FROM cats WHERE 1=1"
	args := []interface{}{}

	if id != "" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
			return
		}
		query += " AND id = " + strconv.Itoa(idInt)
		args = append(args)
	}

	if race != "" {
		// Capitalize the first letter and lowercase the rest of the string
		race = strings.ToLower(race)
		race = strings.Title(race)

		// Validate race against allowed values
		allowedRaces := []string{"Persian", "Maine Coon", "Siamese", "Ragdoll", "Bengal", "Sphynx", "British Shorthair", "Abyssinian", "Scottish Fold", "Birman"}
		var validRace bool
		for _, r := range allowedRaces {
			if race == r {
				validRace = true
				break
			}
		}
		if !validRace {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid race"})
			return
		}
		// Append WHERE clause for race if it's valid
		query += " AND race = '" + race + "'"
		args = append(args)
	}

	if sex != "" {
		// Validate sex against allowed values
		if sex != "male" && sex != "female" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sex"})
			return
		}
		// Append WHERE clause for sex if it's valid
		query += " AND sex = '" + sex + "'"
		args = append(args)
	}

	if hasMatchedStr != "" {
		// Parse hasMatched as boolean
		// hasMatched, err := strconv.ParseBool(hasMatchedStr)
		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hasMatched value"})
		// 	return
		// }
		// Append WHERE clause for hasMatched
		query += " AND has_matched = " + hasMatchedStr
		args = append(args)
	}

	if ageInMonth != "" {
		// Parse ageInMonth filter
		var ageCondition string
		if strings.HasPrefix(ageInMonth, ">") {
			ageCondition = ">"
		} else if strings.HasPrefix(ageInMonth, "<") {
			ageCondition = "<"
		} else if strings.HasPrefix(ageInMonth, "") {
			ageCondition = "="
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ageInMonth filter"})
			return
		}

		ageValue, err := strconv.Atoi(strings.TrimPrefix(ageInMonth, ageCondition))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ageInMonth value"})
			return
		}

		// Append WHERE clause for ageInMonth
		query += " AND age_in_month " + ageCondition + " " + strconv.Itoa(ageValue)
		args = append(args)
	}

	if ownedStr != "" {
		// Parse owned as boolean
		owned, err := strconv.ParseBool(ownedStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owned value"})
			return
		}

		// If owned is true, filter cats with a non-null user_id
		if owned {
			query += " AND deleted_at IS NULL"
		} else {
			// If owned is false, filter cats with a null user_id
			query += " AND deleted_at IS NOT NULL"
		}
	} else {
		deletedAt = " AND deleted_at IS NULL"
		query += deletedAt
		args = append(args)
	}

	if search != "" {
		// Append WHERE clause for search
		query += " AND name LIKE '%" + search + "%'"
		args = append(args)
	}

	if limit != "" && offset != "" {
		query += " ORDER BY created_at DESC LIMIT " + limit + " OFFSET " + offset
		args = append(args)
	}
	fmt.Println(query)

	// Retrieve cats from the database, excluding soft-deleted ones
	rows, err := DB.Query(query, args...)
	if err != nil {
		log.Println("Error retrieving cats:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cats"})
		return
	}

	// Create slice to hold retrieved cats
	cats := []gin.H{}

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
			"id":          strconv.Itoa(cat.ID),
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
	defer DB.Close()
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
	var catDeletedAt sql.NullTime
	var catHasMatched bool
	err = DB.QueryRow("SELECT user_id, deleted_at, has_matched from cats where id=$1", catID).Scan(&catUserId, &catDeletedAt, &catHasMatched)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if catUserId != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not Allowed to Update"})
		return
	}
	if catDeletedAt.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update a soft deleted cat"})
		return
	}
	if catHasMatched {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot Update Cat has matched"})
		return
	}

	var exists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM match_cats WHERE (issuedCatId=$1 OR receiverCatId=$1) AND deleted_at IS NULL)", catID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cat cannot be updated because it is involved in matches"})
		return
	}

	// Bind request body to Cat struct
	var cat struct {
		Name        string   `json:"name" binding:"required,min=1,max=30"`
		Race        string   `json:"race" binding:"required,oneof=Persian 'Maine Coon' Siamese Ragdoll Bengal Sphynx 'British Shorthair' Abyssinian 'Scottish Fold' Birman"`
		Sex         string   `json:"sex" binding:"required,oneof=male female"`
		AgeInMonth  int      `json:"ageInMonth" binding:"required,min=1,max=120082"`
		Description string   `json:"description" binding:"required,min=1,max=200"`
		ImageURLs   []string `json:"imageUrls" binding:"required,min=1,dive,url"`
	}
	fmt.Println(cat)

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
	defer DB.Close()

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

	var exists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM match_cats WHERE (issuedCatId=$1 OR receiverCatId=$1) AND deleted_at IS NULL)", catID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cat cannot be deleted because it is involved in matches"})
		return
	}

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

	// Soft delete: set deleted_at field
	_, err = DB.Exec("UPDATE cats SET deleted_at = NOW() WHERE id = $1", catID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cat"})
		return
	}
	defer DB.Close()

	c.JSON(http.StatusOK, gin.H{"message": "Cat deleted successfully"})
}
