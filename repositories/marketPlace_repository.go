package repositories

import (
	"GameWala-Arcade/models"
	"GameWala-Arcade/utils"
	"database/sql"
	"fmt"
)

type MarketPlaceRepository interface {
	FetchProducts(productType models.ProductType) ([]models.Product, error)
}

type marketPlaceRepository struct {
	db *sql.DB
}

func NewMarketPlaceReposiory(db *sql.DB) *marketPlaceRepository {
	return &marketPlaceRepository{db: db}
}

func (r *marketPlaceRepository) FetchProducts(productType models.ProductType) ([]models.Product, error) {
	utils.LogInfo("Getting products for type: %d", productType)

	rows, err := r.db.Query(`SELECT id, "productName", price, description, units FROM "Products" WHERE "productType" = $1`, productType)
	if err != nil {
		utils.LogError("some error occured while querying db: %v", err)
		return nil, fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	// Slice to hold the result set
	var products []models.Product

	// Iterate through the rows and map to Product struct
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ProductId, &product.Title, &product.Price, &product.Description, &product.TotalUnits); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		products = append(products, product)
	}

	// Check for errors after iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error with row iteration: %v", err)
	}

	return products, nil
}
