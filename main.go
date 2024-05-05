package main

import (
	"CatsSocial/configurations"
	"CatsSocial/controllers"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	DB, err := configurations.DBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
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
	defer DB.Close()
}
