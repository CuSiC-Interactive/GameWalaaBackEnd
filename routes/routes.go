package routes

import (
	"GameWala-Arcade/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine,
	adminConsoleHandler handlers.AdminConsoleHandler,
	playGameHandler handlers.PlayGameHandler,
	handlePaymentHandler handlers.HandlePaymentHandler,
	marketPlaceHandler handlers.MarketPlaceHandler) {
	v1 := router.Group("/api/v1")
	{
		admin := v1.Group("/restricted")
		{
			admin.POST("/signup", adminConsoleHandler.SignUp)
			admin.GET("/login", adminConsoleHandler.Login) //login the admin
			admin.POST("/", adminConsoleHandler.AddGames)  // add games(C)
			//CRUD, R is not there, will be the part of different group.
			// admin.POST("/", adminConsoleHandler.AddGames)
			// admin.PUT("/", adminConsoleHandler.UpdateGames)
			// admin.DELETE("/", adminConsoleHandler.DeleteGames)
		}

		users := v1.Group("")
		{
			users.GET("/games", playGameHandler.GetGamesCatalogue)
			users.POST("/games/status", playGameHandler.SaveGameStatus)
			users.GET("/code-check/:gamecode", playGameHandler.CheckGameCode)
			// users.GET("code-generate", playGameHandler.GenerateCode) // unexposed, not needed
		}

		payment := v1.Group("payment")
		{
			payment.GET("/order/:amount", handlePaymentHandler.CreateOrder)
			payment.POST("/order/details", handlePaymentHandler.SaveOrderDetails)
		}

		shop := v1.Group("/shop")
		{
			shop.GET("/products", marketPlaceHandler.Products)
		}
	}
}
