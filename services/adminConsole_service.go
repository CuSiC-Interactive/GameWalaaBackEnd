package services

import (
	"GameWala-Arcade/models"
	"GameWala-Arcade/utils"
	"GameWala-Arcade/repositories"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AdminConsoleService interface {
	// Authentication Related
	SignUp(user models.AdminCreds) (int, error)
	Login(creds models.AdminCreds) (string, int, error)
	//crud
}

type adminConsoleService struct {
	adminConsoleRepository repositories.AdminConsoleRepository
}

func NewAdminConsoleService(adminConsoleRepository repositories.AdminConsoleRepository) *adminConsoleService {
	return &adminConsoleService{adminConsoleRepository: adminConsoleRepository}
}

func (s *adminConsoleService) Login(creds models.AdminCreds) (string, int, error) {
	utils.LogInfo("Processing login request for email: %s", creds.Email)
	if creds.Password == "" || creds.Email == "" {
		utils.LogError("Login attempt with empty credentials for email: %s", creds.Email)
		return "", 0, fmt.Errorf("Null Arguments passed to service")
	}

	passHash, username, userId, err := s.adminConsoleRepository.Login(creds)

	if err != nil {
		utils.LogError("Login repository error for email %s: %v", creds.Email, err)
		return username, -1, fmt.Errorf("some error occured: %w", err)
	} else if userId > 0 {
		if checkPasswordHash(creds.Password, passHash) {
			return username, userId, nil
		} else {
			utils.LogError("Password mismatch for user ID %d", userId)
			return "existsButPWNotMatched", userId, fmt.Errorf("provided password does not match")
		}
	}
	return "username", 1, fmt.Errorf("user doesn't exist please check username")
}

func (s *adminConsoleService) SignUp(user models.AdminCreds) (int, error) {
	utils.LogInfo("Processing signup request for email: %s", user.Email)
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		utils.LogError("Failed to hash password for email %s: %v", user.Email, err)
		return 0, fmt.Errorf("problem creating the hash of password: %w", err)
	}
	user.Password = hashedPassword
	return s.adminConsoleRepository.CreateUser(user)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
