package controllers

import (
	"CatsSocial/configurations"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func CreateMatch(c *gin.Context) {
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

	// // Cek apakah cat yang dimaksud milik pengguna
	// var ownerID int
	// err = DB.QueryRow("SELECT user_id FROM cats WHERE id = $1", matchRequest.UserCatID).Scan(&ownerID)
	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "User cat not found"})
	// 	return
	// }

	// if ownerID != userID {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "User Cat not belongs to user"})
	// 	return
	// }

	// Cek apakah gender kucing sama
	var userCatSex string
	var matchCatSex string
	err = DB.QueryRow("SELECT sex FROM cats WHERE id = $1", matchRequest.UserCatID).Scan(&userCatSex)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User cat not found"})
		return
	}
	err = DB.QueryRow("SELECT sex FROM cats WHERE id = $1", matchRequest.MatchCatID).Scan(&matchCatSex)
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
	err = DB.QueryRow("SELECT status FROM match_cats WHERE (issuedCatId = $1 AND receiverCatId = $2) OR (issuedCatId = $2 AND receiverCatId = $1) AND deleted_at IS NULL", matchRequest.UserCatID, matchRequest.MatchCatID).Scan(&isMatched)
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
	err = DB.QueryRow("SELECT user_id FROM cats WHERE id = $1", matchRequest.UserCatID).Scan(&userCatOwnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User cat not found"})
		return
	}
	err = DB.QueryRow("SELECT user_id FROM cats WHERE id = $1", matchRequest.MatchCatID).Scan(&matchCatOwnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match cat not found"})
		return
	}
	// if userCatOwnerID == matchCatOwnerID {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Both cats belong to the same owner"})
	// 	return
	// }
	// if userCatOwnerID != userID {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "User Cat not belongs to user"})
	// 	return
	// }

	// Tambahkan permintaan pencocokan kucing ke database
	_, err = DB.Exec("INSERT INTO match_cats (issuedId, issuedCatId, receiverId, receiverCatId, message, status) VALUES ($1, $2, $3, $4, $5, false)",
		userID, matchRequest.UserCatID, matchCatOwnerID, matchRequest.MatchCatID, matchRequest.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add match request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Match request sent successfully"})
	defer DB.Close()
}

func GetMatchRequests(c *gin.Context) {
	// Mendapatkan user ID dari token JWT
	userID, err := configurations.GetUserFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if userID == 0 {
		// mc.issuedId = $1 OR mc.receiverId = $1 AND
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	type CatDetail struct {
		ID          int       `json:"id"`
		Name        string    `json:"name"`
		Race        string    `json:"race"`
		Sex         string    `json:"sex"`
		AgeInMonth  int       `json:"ageInMonth"`
		Description string    `json:"description"`
		ImageURLs   []string  `json:"imageUrls"`
		Status      bool      `json:"hasMatched"`
		CreatedAt   time.Time `json:"createdAt"`
	}
	type User struct {
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"createdAt"`
	}

	// Retrieve match requests from the database
	rows, err := DB.Query("SELECT mc.id, u1.name AS issuedName,  u1.email AS issuedEmail, u1.created_at AS issuedAt,  c1.id AS issuedCatId, c1.name AS issuedCatName, c1.race AS issuedCatRace, c1.sex AS issuedCatSex, c1.age_in_month AS issuedCatAgeInMonth, c1.description AS issuedCatDescription, c1.image_urls AS issuedCatImageUrls, c1.has_matched AS issuedCatStatus, c1.created_at AS issuedCatCreatedAt, c2.id AS receiverCatId, c2.name AS receiverCatName, c2.race AS receiverCatRace, c2.sex AS receiverCatSex, c2.age_in_month AS receiverCatAgeInMonth, c2.description AS receiverCatDescription, c2.image_urls AS receiverCatImageUrls, c2.has_matched AS receiverCatStatus, c2.created_at AS receiverCatCreatedAt, mc.message, mc.created_at FROM match_cats mc INNER JOIN users u1 ON mc.issuedId = u1.id INNER JOIN users u2 ON mc.receiverId = u2.id INNER JOIN cats c1 ON mc.issuedCatId = c1.id INNER JOIN cats c2 ON mc.receiverCatId = c2.id WHERE mc.deleted_at IS NULL ORDER BY mc.created_at DESC")
	if err != nil {
		log.Println("Error retrieving match requests:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match requests"})
		return
	}
	// defer rows.Close()

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
			pq.Array(&matchRequest.UserCatDetail.ImageURLs),
			&matchRequest.UserCatDetail.Status,
			&matchRequest.UserCatDetail.CreatedAt,
			&matchRequest.MatchCatDetail.ID, // Ambil ID dari MatchCatDetail
			&matchRequest.MatchCatDetail.Name,
			&matchRequest.MatchCatDetail.Race,
			&matchRequest.MatchCatDetail.Sex,
			&matchRequest.MatchCatDetail.AgeInMonth,
			&matchRequest.MatchCatDetail.Description,
			pq.Array(&matchRequest.MatchCatDetail.ImageURLs),
			&matchRequest.MatchCatDetail.Status,
			&matchRequest.MatchCatDetail.CreatedAt,
			&matchRequest.Message,
			&matchRequest.CreatedAt,
		)

		if err != nil {
			log.Println("Error scanning row:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match requests"})
			return
		}

		matchRequests = append(matchRequests, gin.H{
			"id": strconv.Itoa(matchRequest.ID), // Convert ID to string
			"issuedBy": gin.H{
				"name":      matchRequest.IssuedBy.Name,
				"email":     matchRequest.IssuedBy.Email,
				"createdAt": matchRequest.IssuedBy.CreatedAt,
			},
			"matchCatDetail": gin.H{
				"id":          strconv.Itoa(matchRequest.MatchCatDetail.ID), // Convert ID to string
				"name":        matchRequest.MatchCatDetail.Name,
				"race":        matchRequest.MatchCatDetail.Race,
				"sex":         matchRequest.MatchCatDetail.Sex,
				"ageInMonth":  matchRequest.MatchCatDetail.AgeInMonth,
				"description": matchRequest.MatchCatDetail.Description,
				"imageUrls":   matchRequest.MatchCatDetail.ImageURLs,
				"status":      matchRequest.MatchCatDetail.Status,
				"createdAt":   matchRequest.MatchCatDetail.CreatedAt,
			},
			"userCatDetail": gin.H{
				"id":          strconv.Itoa(matchRequest.UserCatDetail.ID), // Convert ID to string
				"name":        matchRequest.UserCatDetail.Name,
				"race":        matchRequest.UserCatDetail.Race,
				"sex":         matchRequest.UserCatDetail.Sex,
				"ageInMonth":  matchRequest.UserCatDetail.AgeInMonth,
				"description": matchRequest.UserCatDetail.Description,
				"imageUrls":   matchRequest.UserCatDetail.ImageURLs,
				"status":      matchRequest.UserCatDetail.Status,
				"createdAt":   matchRequest.UserCatDetail.CreatedAt,
			},
			"message":   matchRequest.Message,
			"createdAt": matchRequest.CreatedAt.Format(time.DateTime),
		})
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "data": matchRequests})
	defer DB.Close()
}

func ApproveMatch(c *gin.Context) {
	userID, err := configurations.GetUserFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
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
	var status bool

	var issuedCatId int
	var receiverCatId int
	var deletedAt time.Time
	err = DB.QueryRow("SELECT status, issuedCatId, receiverCatId, deleted_at FROM match_cats WHERE id = $1 AND receiverId = $2", approval.MatchID, userID).Scan(&status, &issuedCatId, &receiverCatId, &deletedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match request"})
		return
	}

	if status {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match Id is no longer valid"})
		return
	}
	if !status && !deletedAt.IsZero() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match Id is no longer valid"})
		return
	}

	// Approve cat matching request
	_, err = DB.Exec("UPDATE match_cats SET status = true WHERE id = $1", approval.MatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve match request"})
		return
	}

	_, err = DB.Exec("UPDATE match_cats SET status = true, deleted_at = NOW() WHERE (issuedCatId = $1 OR receiverCatId = $2) AND id != $3", issuedCatId, receiverCatId, approval.MatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve match request"})
		return
	}
	_, err = DB.Exec("UPDATE cats SET has_matched = true WHERE id = $1 or id = $2", issuedCatId, receiverCatId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve match request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match request approved successfully"})
	defer DB.Close()
}

func RejectMatch(c *gin.Context) {
	userID, err := configurations.GetUserFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
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
	var status bool
	var issuedCatId int
	var receiverCatId int
	var deletedAt time.Time
	err = DB.QueryRow("SELECT status, issuedCatId, receiverCatId, deleted_at FROM match_cats WHERE id = $1 AND receiverId = $2", rejection.MatchID, userID).Scan(&status, &issuedCatId, &receiverCatId, &deletedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match request"})
		return
	}

	if status {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match Id is no longer valid"})
		return
	}
	if !status && !deletedAt.IsZero() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match Id is no longer valid"})
		return
	}

	// Reject cat matching request by updating deleted_at column
	_, err = DB.Exec("UPDATE match_cats SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", rejection.MatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject match request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match request rejected successfully"})
	defer DB.Close()
}

func DeleteMatch(c *gin.Context) {
	userID, err := configurations.GetUserFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Mendapatkan ID pencocokan dari path params
	matchID := c.Param("id")

	// Check if match request exists in the database and the user is the issuer
	var status bool
	var issuedId int
	var issuedCatId int
	var receiverCatId int
	var deletedAt time.Time
	err = DB.QueryRow("SELECT status, issuedId, issuedCatId, receiverCatId, deleted_at FROM match_cats WHERE id = $1 AND receiverId = $2", matchID, userID).Scan(&status, &issuedId, &issuedCatId, &receiverCatId, &deletedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match request"})
		return
	}

	if issuedId != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You Cannot delete someone match"})
		return
	}

	if status {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match Id is already Approved"})
		return
	}
	if !status && !deletedAt.IsZero() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match Id is already Reject"})
		return
	}

	// Delete cat matching request by updating deleted_at column
	_, err = DB.Exec("UPDATE match_cats SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete match request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match request deleted successfully"})
	defer DB.Close()
}
