package handlers

import (
	"GameWala-Arcade/config"
	"GameWala-Arcade/models"
	"GameWala-Arcade/services"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	razorpay "github.com/razorpay/razorpay-go"
	"github.com/redis/go-redis/v9"
)

type HandlePaymentHandler interface {
	CreateOrder(c *gin.Context)
	SaveOrderDetails(c *gin.Context)
	GameLaunchStatus(c *gin.Context)
}

type handlePaymentHandler struct {
	handlePaymentService services.HandlePaymentService
	redisClient          *redis.Client
}

func NewHandlePaymentHandler(paymentService services.HandlePaymentService, redisClient *redis.Client) *handlePaymentHandler {
	// Test Redis connectivity on handler creation
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("[REDIS-ERROR] Could not connect to Redis: %v\n", err)
	} else {
		fmt.Println("[REDIS-DEBUG] Successfully connected to Redis.")
	}
	return &handlePaymentHandler{handlePaymentService: paymentService, redisClient: redisClient}
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
		// Log the actual error for debugging
		fmt.Printf("Razorpay order creation error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("razorpay might be down, please try later.").Error()})
	}
}

func (h *handlePaymentHandler) SaveOrderDetails(c *gin.Context) {
	fmt.Println("[LOGIC-DEBUG] SaveOrderDetails called")
	// Combine PaymentStatus and extra fields in one struct
	type PaymentRequest struct {
		models.PaymentStatus
		UserId   string `json:"userId"`
		TimeLimit int   `json:"timeLimit"`
	}
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("[LOGIC-DEBUG] Invalid payment details format:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment details format provided"})
		return
	}
	if req.UserId == "" || req.TimeLimit == 0 {
		fmt.Printf("[LOGIC-DEBUG] Missing userId or timeLimit in request: userId=%v, timeLimit=%v\n", req.UserId, req.TimeLimit)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userId or timeLimit in request"})
		return
	}

	fmt.Printf("[LOGIC-DEBUG] Received paymentDetails: %+v\n", req.PaymentStatus)
	fmt.Printf("[LOGIC-DEBUG] Received extra: userId=%s, timeLimit=%d\n", req.UserId, req.TimeLimit)

	err := h.handlePaymentService.SaveOrderDetails(req.PaymentStatus)

	if err != nil {
		fmt.Println("[LOGIC-DEBUG] Error saving payment details:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Some error saving payment details. please check logs."})
	} else {
		// Set launch flag and time limit in Redis for this user
		ctx := context.Background()
		key := fmt.Sprintf("game_launch:%s", req.UserId)
		val := fmt.Sprintf("%d", req.TimeLimit)
		fmt.Printf("[LOGIC-DEBUG] Setting launch status: userId=%s, timeLimit=%d\n", req.UserId, req.TimeLimit)
		fmt.Printf("[REDIS-DEBUG] Setting Redis key %s to %s for %d minutes\n", key, val, req.TimeLimit)
		err := h.redisClient.Set(ctx, key, val, time.Duration(req.TimeLimit)*time.Minute).Err()
		if err != nil {
			fmt.Printf("[REDIS-ERROR] Failed to set launch status in Redis: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to set launch status in Redis"})
			return
		}
		fmt.Println("[LOGIC-DEBUG] Launch status set successfully for userId:", req.UserId)
		fmt.Println("[REDIS-DEBUG] Successfully set launch status in Redis")
		c.JSON(http.StatusOK, gin.H{"Success": "Successfully saved order details and launch status."})
	}
}

func (h *handlePaymentHandler) GameLaunchStatus(c *gin.Context) {
	userId := c.Query("userId")
	fmt.Printf("[DEBUG] GameLaunchStatus called for userId=%s\n", userId)
	if userId == "" {
		fmt.Println("[DEBUG] Missing userId in query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userId in query"})
		return
	}
	ctx := context.Background()
	key := fmt.Sprintf("game_launch:%s", userId)
	val, err := h.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		fmt.Printf("[REDIS-DEBUG] Redis key %s not found\n", key)
		c.JSON(http.StatusOK, gin.H{"launch": false, "time_limit": 0})
		return
	} else if err != nil {
		fmt.Printf("[REDIS-ERROR] Error getting Redis key %s: %v\n", key, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get launch status from Redis"})
		return
	}
	fmt.Printf("[REDIS-DEBUG] Redis key %s found, value=%s\n", key, val)
	// If value exists, launch is true
	c.JSON(http.StatusOK, gin.H{"launch": true, "time_limit": val})
}

