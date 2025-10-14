package handlers

import (
	"GameWala-Arcade/models"
	"GameWala-Arcade/services"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type MarketPlaceHandler interface {
	Products(c *gin.Context)
}

type marketPlaceHandler struct {
	marketPlaceService services.MarketPlaceService
}

func NewMarketPlaceHandler(marketPlaceService services.MarketPlaceService) *marketPlaceHandler {
	return &marketPlaceHandler{marketPlaceService: marketPlaceService}
}

// get the products, type will be defined in queryparam, like ?cards, or ?stickers
// Note: This API may need pagination in future.
func (h *marketPlaceHandler) Products(c *gin.Context) {
	requestedType := c.DefaultQuery("type", "")
	productType, err := stringToProductType(requestedType)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("Wrong product type requested '%s'", requestedType).Error()})
	}

	productData, err := h.marketPlaceService.FetchProducts(productType)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Something went wrong '%v',", err).Error()})
	}

	c.JSON(http.StatusOK, gin.H{"Success": productData})
}

func stringToProductType(s string) (models.ProductType, error) {
	switch strings.ToLower(s) {
	case "sticker":
		return 1, nil // sticker
	case "card":
		return 2, nil // card
	default:
		return 0, fmt.Errorf("invalid product type: %s", s)
	}
}
