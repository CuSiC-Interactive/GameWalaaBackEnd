package handlers

import (
	"GameWala-Arcade/models"
	"GameWala-Arcade/services"
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

const minPrice = 10

type PlayGameHandler interface {
	SaveGameStatus(c *gin.Context)
	// I don't know if we need a poller yet, needs to get my hand dirty with Pi to see if we can utlize the inbuilt timer.
	// ChangeGameStatus()
	GetGamesCatalogue(c *gin.Context) // get the games from games table if displayable is true
}

type playGameHandler struct {
	playGameService services.PlayGameService
}

func NewPlayGameHandler(arcadeStoreService services.PlayGameService) *playGameHandler {
	return &playGameHandler{playGameService: arcadeStoreService}
}

func (h *playGameHandler) SaveGameStatus(c *gin.Context) {
	var req models.GameStatus

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if req.GameId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id provided"})
	}

	if isAnyEmpty(req.Code, req.PaymentReference) {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "invalid code, or payment reference id"})
	}

	if req.Price < minPrice {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price is very low, seems like an attempt to play for free or cheap?"})
	}

	if req.PlayTime == nil && req.Levels == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time and Level, both can't be null"})
	}

	res, err := h.playGameService.SaveGameStatus(req)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Either code '%s' or paymentId '%s' already exists", req.Code, req.PaymentReference),
				})
				return
			}
		} else if res == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("there seems to be some error: %w,please save the code %s and try again after some time!!", err, req.Code).Error()})
			return
		} else if res == 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("given the game: %s , price: %d and time: %d doesn't match.", req.Name, req.Price, *req.PlayTime)})
			return
		} else if res == 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("given the game: %s , price: %d and level: %d doesn't match.", req.Name, req.Price, *req.Levels)})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Some unknown error occured, please save the code %s and try again after some time!!", req.Code)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": fmt.Sprintf("You can proceed to play the game, please enter the code '%s' in arcade console.", req.Code)})
}

func (h *playGameHandler) GetGamesCatalogue(c *gin.Context) {
	// var res models.GameResponse
	res, err := h.playGameService.GetGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Some error occurred: %w", err).Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": res})
}
