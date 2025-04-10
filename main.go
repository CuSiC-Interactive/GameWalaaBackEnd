package main

import (
	"GameWala-Arcade/db"
	"GameWala-Arcade/handlers"
	"GameWala-Arcade/repositories"
	"GameWala-Arcade/routes"
	"GameWala-Arcade/services"

	"GameWala-Arcade/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default() // initialize the router for gin.
	config.LoadConfig()     // load the configurations.
	db.Initialize()         // Initlialize the db based on the configs loaded.

	// cors
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:xyz"}, // Allow the frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true, // Allow cookies to be sent with cross-origin requests
	}))

	adminConsoleRepository := repositories.NewAdminConsoleRepository(db.DB)
	adminConsoleService := services.NewAdminConsoleService(adminConsoleRepository)
	adminConsoleHandler := handlers.NewAdminConsoleHandler(adminConsoleService)

	routes.SetupRoutes(router, adminConsoleHandler)

	router.Run("0.0.0.0:8080")
}
