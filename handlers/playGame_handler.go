package handlers

import (
	"GameWala-Arcade/services"
	"GameWala-Arcade/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const minPrice = 10

type PlayGameHandler interface {
	// SaveGameStatus(c *gin.Context)
	GetGamesCatalogue(c *gin.Context)
	// CheckGameCode(c *gin.Context)
	// GenerateCode(c *gin.Context) //this is something logical ughh.
}

type playGameHandler struct {
	playGameService services.PlayGameService
	ArcadeService   services.ArcadeService
}

func NewPlayGameHandler(arcadeStoreService services.PlayGameService, arcadeService services.ArcadeService) *playGameHandler {
	return &playGameHandler{playGameService: arcadeStoreService, ArcadeService: arcadeService}
}

// func (h *playGameHandler) SaveGameStatus(c *gin.Context) {
// 	utils.LogInfo("Received save game status request")
// 	var req models.GameStatus

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		utils.LogError("Invalid game status input: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
// 		return
// 	}

// 	if req.GameId <= 0 {
// 		utils.LogError("Invalid game ID provided: %d", req.GameId)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id provided"})
// 	}

// 	if isAnyEmpty(req.PaymentReference) {
// 		utils.LogError("Missing payment reference for game ID: %d", req.GameId)
// 		c.JSON(http.StatusPaymentRequired, gin.H{"error": "invalid code, or payment reference id"})
// 	}

// 	if req.Price < minPrice {
// 		utils.LogError("Attempt to play with low price: %d (min: %d) for game ID: %d", req.Price, minPrice, req.GameId)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Price is very low, seems like an attempt to play for free or cheap?"})
// 	}

// 	if req.PlayTime == nil && req.Levels == nil {
// 		utils.LogError("Both PlayTime and Levels are null for game ID: %d", req.GameId)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Time and Level, both can't be null"})
// 	}

// 	res, code, err := h.playGameService.SaveGameStatus(req)

// 	if err != nil {
// 		utils.LogError("Error saving game status for game ID %d: %v", req.GameId, err)
// 		var pqErr *pq.Error
// 		if errors.As(err, &pqErr) {
// 			if pqErr.Code == "23505" {
// 				utils.LogError("Either code '%s' or paymentId '%s' already exists", req.Code, req.PaymentReference)
// 				c.JSON(http.StatusBadRequest, gin.H{
// 					"error": fmt.Sprintf("Either code '%s' or paymentId '%s' already exists", req.Code, req.PaymentReference),
// 				})
// 				return
// 			}
// 		} else if res == 0 {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("there seems to be some error: %w,please save the code %s and try again after some time!!", err, req.Code).Error()})
// 			return
// 		} else if res == 2 {
// 			utils.LogError("Given the game: %s , price: %d and time: %d doesn't match.", req.Name, req.Price, *req.PlayTime)
// 			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Given the game: %s , price: %d and time: %d doesn't match.", req.Name, req.Price, *req.PlayTime)})
// 			return
// 		} else if res == 3 {
// 			utils.LogError("Given the game: %s , price: %d and level: %d doesn't match.", req.Name, req.Price, *req.Levels)
// 			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Given the game: %s , price: %d and level: %d doesn't match.", req.Name, req.Price, *req.Levels)})
// 			return
// 		} else {
// 			utils.LogError("something went wrong seems like server issue, error: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Some unknown error occured, please save the code %s and try again after some time!!", req.Code)})
// 			return
// 		}
// 	}

// 	utils.LogInfo("Game status saved successfully for game ID %d with code %s", req.GameId, code)
// 	c.JSON(http.StatusOK, gin.H{"Code": code})
// }

func (h *playGameHandler) GetGamesCatalogue(c *gin.Context) {
	utils.LogInfo("Received request to get games catalogue")
	// var res models.GameResponse
	arcade_id := c.Request.URL.Query().Get("id")
	if arcade_id != "" {
		isValid, err := h.ArcadeService.ValidateArcade(arcade_id)
		if err != nil {
			utils.LogError("Error validating arcade ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Error validating arcade ID: %w", err).Error()})
			return
		}

		if !isValid {
			utils.LogError("Invalid arcade ID provided: %s", arcade_id)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid arcade ID provided"})
			return
		}
	}

	res, err := h.playGameService.GetGames(arcade_id)
	if err != nil {
		utils.LogError("Error fetching games catalogue: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Some error occurred: %w", err).Error()})
		return
	}

	utils.LogInfo("Successfully retrieved games catalogue")
	c.JSON(http.StatusOK, gin.H{"games": res})
}

// func (h *playGameHandler) CheckGameCode(c *gin.Context) {
// 	code := c.Param("gamecode")

// 	if isAnyEmpty(code) {
// 		utils.LogError("Empty code provided")
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "code is empty or null"})
// 		return
// 	}

// 	details, err := h.playGameService.CheckGameCode(code)
// 	if err != nil {

// 		if strings.Contains(err.Error(), "Scan error") {
// 			utils.LogError("scan error occurred (more likely wrong code entered): %v", err)
// 			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("Wrong error code entered: '%s'", code).Error()})
// 			return
// 		}

// 		utils.LogError("something went wrong please have a look 👀: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "The code you've entered doesn't hold any entry yet, please make sure you have entered the right code."})
// 		return
// 	}

// 	// deprecated portion, still keeping it for a while jic
// 	// if details.IsPlayed == true {
// 	// 	c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("This code has already been played: '%s', please get a new code.", err).Error()})
// 	// 	return
// 	// }

// 	// utils.LogInfo("Code Verified: %s", code)
// 	c.JSON(http.StatusOK, gin.H{"success": details})
// 	return
// }

// func (h *playGameHandler) GenerateCode(c *gin.Context) {

// 	code, err := h.playGameService.GenerateCode()
// 	if err != nil {
// 		utils.LogInfo("Something went wrong: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Generated code is: %v", err).Error()})
// 		return
// 	}

// 	utils.LogInfo("Code is Generated Successfully: %s", code)
// 	c.JSON(http.StatusOK, gin.H{"success": fmt.Sprintf("Generated code is: %s", code)})
// }
