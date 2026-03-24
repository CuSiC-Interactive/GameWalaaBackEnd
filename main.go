package main

import (
	"GameWala-Arcade/db"
	"GameWala-Arcade/handlers"
	"GameWala-Arcade/repositories"
	"GameWala-Arcade/routes"
	"GameWala-Arcade/services"
	"log"

	"GameWala-Arcade/config"
	"GameWala-Arcade/utils"

	mqtt "GameWala-Arcade/utils/mqtt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	// redisStore := redis.NewClient(&redis.Options{
	// 	Addr:     "127.0.0.1:55000",
	// 	Password: "", // No password set
	// 	DB:       0,  // Use default DB
	// })

	// _, err := redisStore.Ping(context.Background()).Result()
	// if err != nil {
	// 	utils.LogError("could not connect to Redis, error: %v", err)
	// 	log.Fatalf("Could not connect to Redis: %v", err)
	// }

	// Initialize MQTT connections once
	mqttService, err := mqtt.NewMQTTService(
		"tcp://localhost:1883",
		"backend-server",
	)
	if err != nil {
		log.Fatal(err)
	}

	// cors
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:xyz"}, // Allow the frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true, // Allow cookies to be sent with cross-origin requests
	}))

	arcadeRepository := repositories.NewArcadeRepository(db.DB)
	ArcadeService := services.NewArcadeService(arcadeRepository)

	adminConsoleRepository := repositories.NewAdminConsoleRepository(db.DB)
	adminConsoleService := services.NewAdminConsoleService(adminConsoleRepository)
	adminConsoleHandler := handlers.NewAdminConsoleHandler(adminConsoleService)

	playGameRepository := repositories.NewPlayGameReposiory(db.DB)
	playGameService := services.NewPlayGameService(playGameRepository)
	handlePublishService := services.NewConnectionToBrokerService(mqttService, playGameRepository)
	playGameHandler := handlers.NewPlayGameHandler(playGameService, ArcadeService)

	handlePaymentRepository := repositories.NewHandlePaymentReposiory(db.DB)
	handlePaymentService := services.NewHandlePaymentService(handlePaymentRepository, playGameRepository)
	handlePaymentHandler := handlers.NewHandlePaymentHandler(handlePaymentService, handlePublishService, ArcadeService)

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
