package services

import (
	"GameWala-Arcade/models"
	"GameWala-Arcade/repositories"
	"GameWala-Arcade/utils"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var maxTimeForLevelBoundedGame = uint16(120)

const staticStartingCode = "ABXYSO"

type PlayGameService interface {
	SaveGameStatus(status models.GameStatus) (int, string, error)
	GetGames() ([]models.GameResponse, error)
	CheckGameCode(code string) (models.GameDetails, error) // arcade will hit this api
	GenerateCode() (string, error)
}

type playGameService struct {
	playGameRepository repositories.PlayGameRepository
	redisClient        *redis.Client
}

func NewPlayGameService(playGameRepository repositories.PlayGameRepository,
	redisClient *redis.Client) *playGameService {
	return &playGameService{playGameRepository: playGameRepository, redisClient: redisClient}
}

func (s *playGameService) SaveGameStatus(status models.GameStatus) (int, string, error) {
	utils.LogInfo("Processing save game status for game ID %d", status.GameId)

	if status.IsTimed && status.PlayTime != nil {
		err := s.validateTimeAndPrice(status.GameId, status.Price, status.PlayTime)

		if err != nil {
			utils.LogError("Time and price validation failed for game ID %d: %v", status.GameId, err)
			return 2, "", err // 2 means, time and price didn't match (convert this to enum later)
		}
	} else {
		err := s.validateLevelsAndPrice(status.GameId, status.Price, status.Levels)

		if err != nil {
			utils.LogError("Level and price validation failed for game ID %d: %v", status.GameId, err)
			return 3, "", err // 3 means, level and price didn't match (convert this to enum later)
		}
	}

	code, err := s.GenerateCode()
	status.Code = code
	res, err := s.playGameRepository.SaveGameStatus(status)

	if err != nil {
		utils.LogError("Failed to save game status for game ID %d: %v", status.GameId, err)
		return 0, "", err
	}

	if res == 1 {
		utils.LogInfo("Successfully saved game status for game ID %d", status.GameId)
		return 1, code, err
	}
	return 0, "", err
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

func (s *playGameService) CheckGameCode(code string) (models.GameDetails, error) {
	if code == "" {
		utils.LogError("empty code in service layer? something's fishy 🐠")
		return models.GameDetails{}, fmt.Errorf("Code is empty")
	}

	status, err := s.playGameRepository.CheckGameCode(code)

	if err != nil {
		utils.LogError("Something went wrong... hmm BL layer, kinda sv issue?, err: %s", err)
		return status, err
	}

	return status, err
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

func (s *playGameService) GenerateCode() (string, error) {
	ctx := context.Background()
	latestCode, err := s.redisClient.Get(ctx, "latest_arcade_code").Result()

	if err == redis.Nil {
		latestCode = staticStartingCode                             // starting code
		s.redisClient.Set(ctx, "latest_arcade_code", latestCode, 0) // 0 for no expiration
		return latestCode, err
	}

	newCode := getNextConsecutiveCode(latestCode)
	s.redisClient.Set(ctx, "latest_arcade_code", newCode, 0) // 0 for no expiration
	return newCode, nil
}

func getNextConsecutiveCode(code string) string {
	charset := []rune{'A', 'B', 'O', 'S', 'X', 'Y'}
	base := len(charset)
	runes := []rune(code)
	n := len(runes)

	carry := 1
	for i := n - 1; i >= 0; i-- {
		if carry == 0 {
			break
		}
		idx := indexOf(charset, runes[i])
		if idx == -1 {
			idx = 0
		}
		newIdx := (idx + carry) % base
		carry = (idx + carry) / base
		runes[i] = charset[newIdx]
	}
	return string(runes)
}

func indexOf(slice []rune, r rune) int {
	for i, v := range slice {
		if v == r {
			return i
		}
	}
	return -1
}
