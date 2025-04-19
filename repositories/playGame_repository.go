package repositories

import (
	// "GameWala-Arcade/models"
	"GameWala-Arcade/models"
	"GameWala-Arcade/utils"
	"database/sql"
	"fmt"
)

type PlayGameRepository interface {
	SaveGameStatus(status models.GameStatus) (int, error)
	GetGames() ([]models.GameResponse, error)
	FetchPrices() (models.PriceMap, error)
	CheckGameCode(code string) (bool, *uint16, error)
	ValidateTimeAndPrice(gameId uint16, price uint16, playTime *uint16) error
	ValidateLevelsAndPrice(gameId uint16, price uint16, levels *uint8) error
}

type playGameRepository struct {
	db *sql.DB
}

func NewPlayGameReposiory(db *sql.DB) *playGameRepository {
	return &playGameRepository{db: db}
}

func (r *playGameRepository) SaveGameStatus(status models.GameStatus) (int, error) {
	utils.LogInfo("Saving game status to database for game ID %d", status.GameId)

	// Prepare the call to the stored procedure
	stmt, err := r.db.Prepare("SELECT func_InsertGameStatus($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		utils.LogError("Failed to prepare save game status statement: %v", err)
		return 0, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(status.GameId, status.Name, status.IsPlayed, status.Price,
		status.PlayTime, status.Levels, status.PaymentReference, status.Code)

	if err != nil {
		utils.LogError("Failed to execute save game status for game ID %d: %v", status.GameId, err)
		return 0, fmt.Errorf("error executing function: %w", err)
	}

	utils.LogInfo("Successfully saved game status for game ID %d", status.GameId)
	return 1, nil
}

func (r *playGameRepository) ValidateTimeAndPrice(gameId uint16, price uint16, playTime *uint16) error {
	utils.LogInfo("Validating time and price in database for game ID %d", gameId)
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
		utils.LogError("Invalid time and price combination for game ID %d: price=%d, time=%d", gameId, price, *playTime)
		return fmt.Errorf("wrong combination of price and time provided %w", err)
	}

	return nil
}

func (r *playGameRepository) ValidateLevelsAndPrice(gameId uint16, price uint16, levels *uint8) error {
	utils.LogInfo("Validating levels and price in database for game ID %d", gameId)
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
		utils.LogError("Invalid level and price combination for game ID %d: price=%d, level=%d", gameId, price, *levels)
		return fmt.Errorf("wrong combination of time and level provided %w", err)
	}

	return nil
}

func (r *playGameRepository) GetGames() ([]models.GameResponse, error) {
	utils.LogInfo("Fetching all games from database")

	rows, err := r.db.Query("Select * from func_GetGamesForUsers()")

	if err != nil {
		utils.LogError("Failed to fetch games from database: %v", err)
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

func (r *playGameRepository) CheckGameCode(code string) (bool, *uint16, error) {

	var defaultTime = uint16(0)
	stmt, err := r.db.Prepare("SELECT func_CheckGameCode($1)")

	if err != nil {
		utils.LogError("Failed to prepare save game status statement: %v", err)
		return true, &defaultTime, fmt.Errorf("error preparing statement: %w", err)
	}

	defer stmt.Close()
	var status bool = false
	var time *uint16
	err = stmt.QueryRow(code).Scan(&status, &time)

	if err != nil {
		utils.LogError("Some error occured something wrong with DB? err: %v", err)
		return status, &defaultTime, err
	}

	return status, time, err //if true, then need to implement redis queue to make it false after the time, if timebounded.
}
