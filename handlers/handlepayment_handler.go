package handlers

import (
	"GameWala-Arcade/config"
	"GameWala-Arcade/models"
	"GameWala-Arcade/services"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"

	"GameWala-Arcade/utils"

	"github.com/gin-gonic/gin"
	razorpay "github.com/razorpay/razorpay-go"
)

type HandlePaymentHandler interface {
	CreateOrder(c *gin.Context)
	SaveOrderDetails(c *gin.Context)
}

type handlePaymentHandler struct {
	handlePaymentService services.HandlePaymentService
	handlePublishService services.ConnectionToBrokerService
	ArcadeService        services.ArcadeService
}

func NewHandlePaymentHandler(paymentService services.HandlePaymentService,
	publishService services.ConnectionToBrokerService,
	arcadeService services.ArcadeService) *handlePaymentHandler {
	return &handlePaymentHandler{
		handlePaymentService: paymentService,
		handlePublishService: publishService,
		ArcadeService:        arcadeService,
	}
}

func (h *handlePaymentHandler) CreateOrder(c *gin.Context) {
	arcade_id := c.Param("arcade_id")
	if arcade_id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arcade ID is required"})
		return
	}
	isvalid, err := h.ArcadeService.ValidateArcade(arcade_id)

	if !isvalid || err != nil {
		if err != nil {
			utils.LogError("Error validating arcade ID '%s': %v", arcade_id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Error validating arcade ID: %w", err).Error()})
			return

		}

		utils.LogError("Invalid arcade ID provided: '%s'", arcade_id)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Arcade ID"})
		return
	}

	amount := c.Param("amount")
	amount_inr, _ := strconv.Atoi(amount)
	client := razorpay.NewClient(config.GetString("key_id"), config.GetString("key_secret"))
	receipt := fmt.Sprintf("txn_%d", time.Now().Unix())

	data := map[string]interface{}{
		"amount":   amount_inr,
		"currency": "INR",
		"receipt":  receipt}

	body, err := client.Order.Create(data, map[string]string{}) // 2nd param optional
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"details": body})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("razorpay might be down, please try later.").Error()})
	}
}

func (h *handlePaymentHandler) SaveOrderDetails(c *gin.Context) {
	var paymentAndGameStatus models.PaymentAndGameStatus
	if err := c.ShouldBindJSON(&paymentAndGameStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment and game status details format provided"})
		return
	}
	paymentDetails := paymentAndGameStatus.PaymentDetails
	gameStatus := paymentAndGameStatus.GameStatus

	var arcadeId = gameStatus.ArcadeId
	if arcadeId == "" {
		// least likely to occur as its verified before, but just to be safe. If it occurs that means
		// the payment is needs to be refunded, we will look at this case later.
		// let's continue with happy flow for now, and log this case for future reference.
		utils.LogError("Arcade id is missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arcade ID is required and must be valid"})
		return
	}

	err := h.handlePaymentService.SaveOrderDetails(paymentDetails)
	res, err := h.handlePaymentService.SaveGameStatus(gameStatus)

	handleGameStatus(c, res, err, gameStatus.GameId, gameStatus.Name, gameStatus.Price, gameStatus.PlayTime, gameStatus.Levels, gameStatus.PaymentReference)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Some error saving payment details. please check logs."})
	} else {
		err := h.handlePublishService.PublishMessage(arcadeId, gameStatus)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Payment details saved, but failed to publish message to broker. Please check logs."})
			utils.LogError("could not connect to Redis, error: %v", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"Success: ": "Successfully saved order details and published game details to broker."})
	}
}

func handleGameStatus(c *gin.Context,
	res int,
	err error,
	gameId uint16,
	name string,
	price uint16,
	playTime *uint16,
	levels *uint8,
	paymentReference string) {

	if err != nil {
		utils.LogError("Error saving game status for game ID %d: %v", gameId, err)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				utils.LogError("paymentId '%s' already exists", paymentReference)
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("paymentId '%s' already exists", paymentReference),
				})
				return
			}
		} else if res == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("there seems to be some error: %w,please save the payment reference %s and try again after some time!!", err, paymentReference).Error()})
			return
		} else if res == 2 {
			utils.LogError("Given the game: %s , price: %d and time: %d doesn't match.", name, price, *playTime)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Given the game: %s , price: %d and time: %d doesn't match.", name, price, *playTime)})
			return
		} else if res == 3 {
			utils.LogError("Given the game: %s , price: %d and level: %d doesn't match.", name, price, *levels)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Given the game: %s , price: %d and level: %d doesn't match.", name, price, *levels)})
			return
		}
	}
}
