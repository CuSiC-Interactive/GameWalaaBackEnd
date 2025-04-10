package models

type GameData struct {
	Name      string `json:"name"`
	Price     int    `json:"price"`
	PlayTime  int    `json:"playtime"` // (in minutes)
	Thumbnail string `json:"thumbnail"`
}
