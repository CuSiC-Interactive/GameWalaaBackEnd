package utils

import (
	"GameWala-Arcade/config"
	"fmt"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(config.GetString("secretyKey"))

func CreateToken(username string, id int) (string, error) {
	// Creating a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     username,
		"user_id": id,                                     // Subject (user identifier)
		"iss":     "GameWala",                             // Issuer
		"exp":     time.Now().Add(time.Hour * 168).Unix(), // Expiration time
		"iat":     time.Now().Unix(),                      // Issued at
	})

	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	fmt.Printf("Token claims added: %+v\n", claims)
	return tokenString, nil
}

func AuthenticateMiddleware(c *gin.Context) {
	tokenString, err := c.Cookie("token")
	if err != nil {
		fmt.Println("Token missing in cookie")
		c.JSON(http.StatusUnauthorized, "Token is missing")
		c.Abort()
		return
	}

	// Verify the token
	token, err := verifyToken(tokenString)
	if err != nil {
		fmt.Printf("Token verification failed: %v", err)
		c.JSON(http.StatusUnauthorized, "Token verification failed")
		c.Abort()
		return
	}

	fmt.Printf("Token verified successfully. Claims: %+v", token.Claims)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

	userID, exists := claims["user_id"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in token"})
		c.Abort()
		return
	}

	c.Set("user_id", int(userID))

	c.Next()
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func CheckCookies(c *gin.Context) int {
	UserId, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return 0
	}
	return UserId.(int)
}
