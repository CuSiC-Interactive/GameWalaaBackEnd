package services

import (
	"GameWala-Arcade/models"
	"GameWala-Arcade/repositories"
	"GameWala-Arcade/utils"
)

type PlayGameService interface {
	SaveGameStatus(status models.GameStatus) (int, error)
	GetGames() ([]models.GameResponse, error)
}

type playGameService struct {
	playGameRepository repositories.PlayGameRepository
}

func NewPlayGameService(playGameRepository repositories.PlayGameRepository) *playGameService {
	return &playGameService{playGameRepository: playGameRepository}
}

func (s *playGameService) SaveGameStatus(status models.GameStatus) (int, error) {
	utils.LogInfo("Processing save game status for game ID %d", status.GameId)

	if status.IsTimed && status.PlayTime != nil {
		err := s.validateTimeAndPrice(status.GameId, status.Price, status.PlayTime)

		if err != nil {
			utils.LogError("Time and price validation failed for game ID %d: %v", status.GameId, err)
			return 2, err // 2 means, time and price didn't match (convert this to enum later)
		}
	} else {
		err := s.validateLevelsAndPrice(status.GameId, status.Price, status.Levels)

		if err != nil {
			utils.LogError("Level and price validation failed for game ID %d: %v", status.GameId, err)
			return 3, err // 3 means, level and price didn't match (convert this to enum later)
		}
	}

	res, err := s.playGameRepository.SaveGameStatus(status)

	if err != nil {
		utils.LogError("Failed to save game status for game ID %d: %v", status.GameId, err)
		return 0, err
	}

	utils.LogInfo("Successfully saved game status for game ID %d", status.GameId)
	return res, err
}

func (s *playGameService) GetGames() ([]models.GameResponse, error) {
	utils.LogInfo("Fetching all games from service")
	games, err := s.playGameRepository.GetGames()

	if err != nil {
		utils.LogError("Failed to fetch games: %v", err)
		return nil, err
	}

	prices, err := s.playGameRepository.FetchPrices()

	for game := 0; game < len(games); game++ {
		currId := games[game].GameId
		if len(prices.TimeMap[currId]) > 0 {
			games[game].Price.ByTime = append(games[game].Price.ByTime, prices.TimeMap[currId]...)
		} else if len(prices.LevelMap[currId]) > 0 {
			games[game].Price.ByLevel = append(games[game].Price.ByLevel, prices.LevelMap[currId]...)
		}
	}

	utils.LogInfo("Successfully fetched %d games", len(games))
	return games, nil
}

func (s *playGameService) validateTimeAndPrice(gameId uint16, price uint16, playTime *uint16) error {
	utils.LogInfo("Validating time and price for game ID %d: price=%d, time=%d", gameId, price, *playTime)
	//call db to cheeck if time and price match with the feeded value.
	err := s.playGameRepository.ValidateTimeAndPrice(gameId, price, playTime)

	return err
}

func (s *playGameService) validateLevelsAndPrice(gameId uint16, price uint16, levels *uint8) error {
	utils.LogInfo("Validating levels and price for game ID %d: price=%d, level=%d", gameId, price, *levels)
	//call db to cheeck if level and price match with the feeded value.
	err := s.playGameRepository.ValidateLevelsAndPrice(gameId, price, levels)

	return err
}
