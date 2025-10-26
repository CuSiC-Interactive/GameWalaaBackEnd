package main

import (
	"GameWala-Arcade/db"
	"GameWala-Arcade/handlers"
	"GameWala-Arcade/repositories"
	"GameWala-Arcade/routes"
	"GameWala-Arcade/services"
	"context"
	"log"

	"GameWala-Arcade/config"
	"GameWala-Arcade/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Initialize logger
	if err := utils.InitLogger(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer utils.CloseLogger()

	router := gin.Default() // initialize the router for gin.
	utils.LogInfo("Starting GameWala-Arcade server...")
	config.LoadConfig() // load the configurations.
	db.Initialize()     // Initlialize the db based on the configs loaded.

	redisStore := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Changed from 55003 to 6379
		Password: "", // No password set
		DB:       0,  // Use default DB
	})

	_, err := redisStore.Ping(context.Background()).Result()
	if err != nil {
		utils.LogError("could not connect to Redis, error: %v", err)
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	// cors
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173" , "http://10.0.100.39:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	adminConsoleRepository := repositories.NewAdminConsoleRepository(db.DB)
	adminConsoleService := services.NewAdminConsoleService(adminConsoleRepository)
	adminConsoleHandler := handlers.NewAdminConsoleHandler(adminConsoleService)

	playGameRespository := repositories.NewPlayGameReposiory(db.DB)
	playGameService := services.NewPlayGameService(playGameRespository, redisStore)
	playGameHandler := handlers.NewPlayGameHandler(playGameService)

	handlePaymentRepository := repositories.NewHandlePaymentReposiory(db.DB)
	handlePaymentService := services.NewHandlePaymentService(handlePaymentRepository)
	handlePaymentHandler := handlers.NewHandlePaymentHandler(handlePaymentService)

	marketPlaceRepository := repositories.NewMarketPlaceReposiory(db.DB)
	marketPlaceService := services.NewMarketPlaceService(marketPlaceRepository)
	marketPlaceHandler := handlers.NewMarketPlaceHandler(marketPlaceService)

	routes.SetupRoutes(
		router,
		adminConsoleHandler,
		playGameHandler,
		handlePaymentHandler,
		marketPlaceHandler)

	utils.LogInfo("Server starting on 0.0.0.0:8080")
	if err := router.Run("0.0.0.0:8080"); err != nil {
		utils.LogError("Server failed to start: %v", err)
		panic(err)
	}
}
