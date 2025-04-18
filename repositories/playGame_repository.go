package repositories

import (
	// "GameWala-Arcade/models"
	"GameWala-Arcade/models"
	"database/sql"
	"fmt"
)

type PlayGameRepository interface {
	SaveGameStatus(status models.GameStatus) (int, error)
	ValidateTimeAndPrice(gameId uint16, price uint16, playTime *uint16) error
	ValidateLevelsAndPrice(gameId uint16, price uint16, levels *uint8) error
	GetGames() ([]models.GameResponse, error)
	FetchPrices() (models.PriceMap, error)
}

type playGameRepository struct {
	db *sql.DB
}

func NewPlayGameReposiory(db *sql.DB) *playGameRepository {
	return &playGameRepository{db: db}
}

func (r *playGameRepository) SaveGameStatus(status models.GameStatus) (int, error) {

	// Prepare the call to the stored procedure
	stmt, err := r.db.Prepare("SELECT func_InsertGameStatus($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		return 0, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(status.GameId, status.Name, status.IsPlayed, status.Price,
		status.PlayTime, status.Levels, status.PaymentReference, status.Code)

	if err != nil {
		return 0, fmt.Errorf("error executing function: %w", err)
	}

	return 1, nil
}

func (r *playGameRepository) ValidateTimeAndPrice(gameId uint16, price uint16, playTime *uint16) error {
	stmt, err := r.db.Prepare("Select func_ValidateTimeAndPice($1, $2, $3)")

	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}

	defer stmt.Close()

	var exists bool
	err = stmt.QueryRow(gameId, price, playTime).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("wrong combination of price and time provided %w", err)
	}

	return nil
}

func (r *playGameRepository) ValidateLevelsAndPrice(gameId uint16, price uint16, levels *uint8) error {
	stmt, err := r.db.Prepare("Select func_ValidateLevelsAndPrice($1, $2, $3)")

	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}

	defer stmt.Close()

	var exists bool
	err = stmt.QueryRow(gameId, price, levels).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("wrong combination of time and level provided %w", err)
	}

	return nil
}

func (r *playGameRepository) GetGames() ([]models.GameResponse, error) {

	rows, err := r.db.Query("Select * from func_GetGamesForUsers()")

	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	defer rows.Close()
	var games []models.GameResponse

	for rows.Next() {
		var game models.GameResponse

		err := rows.Scan(&game.GameId, &game.Name, &game.Thumbnail)
		if err != nil {
			return nil, fmt.Errorf("error fetching games: %w", err)
		}
		games = append(games, game)
	}

	return games, nil
}

func (r *playGameRepository) FetchPrices() (models.PriceMap, error) {

	var price models.PriceMap

	rows, err := r.db.Query("SELECT * FROM func_GetGamesPrices()")

	if err != nil {
		return price, fmt.Errorf("query error: %w", err)
	}

	defer rows.Close()

	timePriceMap := make(map[uint16][]models.TimePrice)
	levelPriceMap := make(map[uint16][]models.LevelPrice)

	for rows.Next() {
		var gp models.GamePrice
		if err := rows.Scan(&gp.ItemType, &gp.Label, &gp.Price, &gp.Id); err != nil {
			return price, fmt.Errorf("scan error: %w", err)
		}

		switch gp.ItemType {
		case "time":
			timePriceMap[gp.Id] = append(timePriceMap[gp.Id], models.TimePrice{
				Time:  gp.Label,
				Price: gp.Price,
			})
		case "level":
			levelPriceMap[gp.Id] = append(levelPriceMap[gp.Id], models.LevelPrice{
				Level: gp.Label,
				Price: gp.Price,
			})
		}
	}

	price.TimeMap = timePriceMap
	price.LevelMap = levelPriceMap
	return price, nil

}
