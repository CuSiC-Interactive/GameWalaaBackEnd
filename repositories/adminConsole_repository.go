package repositories

import (
	"GameWala-Arcade/models"
	"GameWala-Arcade/utils"
	"database/sql"
	"fmt"
)

type AdminConsoleRepository interface {
	// Authentication related.
	CreateUser(user models.AdminCreds) (int, error)
	Login(creds models.AdminCreds) (string, string, int, error)

	// CRUD
	GetGames() models.GameData
}

type adminConsoleRepository struct {
	db *sql.DB
}

func NewAdminConsoleRepository(db *sql.DB) *adminConsoleRepository {
	return &adminConsoleRepository{db: db}
}

func (r *adminConsoleRepository) CreateUser(user models.AdminCreds) (int, error) {
	utils.LogInfo("Creating new admin user in database: %s", user.Email)
	var userId int

	// Prepare the call to the stored procedure
	stmt, err := r.db.Prepare("SELECT func_InsertUser($1, $2, $3)")
	if err != nil {
		utils.LogError("Failed to prepare create user statement: %v", err)
		return 0, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	// Retrieve the OUT parameter value
	err = stmt.QueryRow(user.Username, user.Email, user.Password).Scan(&userId)
	if err != nil {
		utils.LogError("Failed to execute create user function for email %s: %v", user.Email, err)
		return 0, fmt.Errorf("error executing function: %w", err)
	}

	utils.LogInfo("Successfully created admin user with ID %d", userId)
	return userId, nil
}

func (r *adminConsoleRepository) Login(creds models.AdminCreds) (string, string, int, error) {
	utils.LogInfo("Fetching login data for email: %s", creds.Email)
	var passwordHash string
	var username string
	userId := 0
	stmt, err := r.db.Prepare("Select * From func_getAdminLoginData($1)")
	if err != nil {
		utils.LogError("Failed to prepare login statement: %v", err)
		return passwordHash, username, userId, fmt.Errorf("error executing function: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(creds.Email).Scan(&passwordHash, &username, &userId)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.LogError("No user found for email: %s", creds.Email)
			return passwordHash, username, userId, err
		}
		return passwordHash, username, userId, fmt.Errorf("error executing function: %w", err)
	}
	utils.LogInfo("Successfully fetched login data for user ID %d", userId)
	return passwordHash, username, userId, nil
}

func (r *adminConsoleRepository) GetGames() models.GameData {
	var x models.GameData
	return x
}
