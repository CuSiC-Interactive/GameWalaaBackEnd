package handlers

import (
	"GameWala-Arcade/services"

	"GameWala-Arcade/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	jwtutil "GameWala-Arcade/utils"

	utils "GameWala-Arcade/utils"
)

type AdminConsoleHandler interface {
	SignUp(c *gin.Context)
	Login(c *gin.Context) // login for admin.

	AddGames(c *gin.Context) // Add games
	// GetGames(c *gin.Context) // get for admin (it's different)
	// DeleteGames(c *gin.Context) //update
	// UpdateGames(c *gin.Context) //delete

}

type adminConsoleHandler struct {
	adminConsoleService services.AdminConsoleService
}

const passwordNotMatched = "existsButPWNotMatched"

func NewAdminConsoleHandler(adminConsoleService services.AdminConsoleService) *adminConsoleHandler {
	return &adminConsoleHandler{adminConsoleService: adminConsoleService}
}

func (h *adminConsoleHandler) SignUp(c *gin.Context) {
	utils.LogInfo("Received admin signup request")

	var user models.AdminCreds
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.LogError("Invalid signup input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if isAnyEmpty(user.Username, user.Email, user.Password) {
		utils.LogError("Empty required fields in signup for user: %s", user.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input, either of the required param is empty"})
		return
	}

	userId, err := h.adminConsoleService.SignUp(user)
	if err != nil {
		utils.LogError("Failed to signup admin: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := fmt.Sprintf("User registered successfully as admin with id %d", userId)
	utils.LogInfo("Admin signup successful: %s (ID: %d)", user.Email, userId)
	c.JSON(http.StatusOK, gin.H{"message": message})
}

func (h *adminConsoleHandler) Login(c *gin.Context) {
	utils.LogInfo("Received admin login request")
	var admin models.AdminCreds
	if err := c.ShouldBindJSON(&admin); err != nil {
		utils.LogError("Invalid login input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if isAnyEmpty(admin.Email, admin.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input, either of the required param is empty"})
	}

	username, adminId, err := h.adminConsoleService.Login(admin)

	if adminId <= 0 && err != nil {
		utils.LogError("Failed login attempt for unregistered admin: %s", admin.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin not registered, are you certain you are the admin? 🤨"})
		return
	} else if adminId > 0 && username == passwordNotMatched {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Wrong password entered %s", err)})
		return
	}

	tokenString, err := jwtutil.CreateToken(username, adminId)
	if err != nil {
		utils.LogError("Failed to create JWT token for admin %s: %v", username, err)
		c.String(http.StatusInternalServerError, "Error creating the authentication token, please try again. maybe servers are down.")
	}

	c.SetCookie(
		"token",
		tokenString,
		36000,
		"/",
		"localhost",
		false, //make sure to make it true later in https
		true)
	utils.LogInfo("Admin login successful: %s (ID: %d)", username, adminId)
	c.JSON(http.StatusOK, gin.H{"name": username, "admin Id": adminId, "message": "Welcome admin!!"})
}

// add jwt middleware in route.
func (h *adminConsoleHandler) AddGames(c *gin.Context) {

}

// private methods
func isAnyEmpty(strings ...string) bool {
	for _, str := range strings {
		if str == "" {
			return true
		}
	}
	return false
}
