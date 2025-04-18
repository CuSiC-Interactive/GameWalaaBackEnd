package models

import (
	"time"
)

type GameData struct {
	Name      string `json:"name"`
	Price     uint8  `json:"price"`
	PlayTime  uint8  `json:"playtime"` // (in minutes)
	Thumbnail string `json:"thumbnail"`
}

type GameStatus struct {
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	GameId           uint16    `json:"gameId"`
	IsTimed          bool      `json:"isTimed"`
	Price            uint16    `json:"price"`
	PlayTime         *uint16   `json:"playTime"`
	Levels           *uint8    `json:"levels"`
	TimeStamp        time.Time `json:"currentTime"`
	IsPlayed         bool      `json:"played"`
	PaymentReference string    `json:"paymentId"`
}

type GameResponse struct {
	Name      string
	GameId    uint16
	Price     Price
	Thumbnail *string
}

type GamePrice struct {
	Id       uint16
	ItemType string
	Label    uint16
	Price    uint16
}

type Price struct {
	ByTime  []TimePrice
	ByLevel []LevelPrice
}

type PriceMap struct {
	TimeMap  map[uint16][]TimePrice
	LevelMap map[uint16][]LevelPrice
}

type TimePrice struct {
	Time  uint16
	Price uint16
}

type LevelPrice struct {
	Level uint16
	Price uint16
}
