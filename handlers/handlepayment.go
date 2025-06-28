package handlers

import (
	"GameWala-Arcade/config"
	"GameWala-Arcade/services"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	razorpay "github.com/razorpay/razorpay-go"
)

type HandlePaymentHandler interface {
	CreateOrder(c *gin.Context)
}

type handlePaymentHandler struct {
	handlePaymentService services.HandlePaymentService
}

func NewHandlePaymentHandler(paymentService services.HandlePaymentService) *handlePaymentHandler {
	return &handlePaymentHandler{handlePaymentService: paymentService}
}

func (h *handlePaymentHandler) CreateOrder(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("razorpay might be down: %w amount: %d", err, amount_inr).Error()})
	}
}
