package models

type Product struct {
	ProductId   int32    `json:"productid"`
	Price       int32    `json:"price"`
	Description string   `json:"description"`
	TotalUnits  int8     `json:"totalunits"`
	Title       string   `json:"title"`
	CoverImage  string   `json:"coverImage"`
	Images      []string `json:"images"`
}

type ProductType int

const (
	Sticker ProductType = iota + 1
	Card
)
