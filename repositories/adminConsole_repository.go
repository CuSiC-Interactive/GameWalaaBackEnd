package repositories

import (
	"GameWala-Arcade/models"
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
	var userId int

	// Prepare the call to the stored procedure
	stmt, err := r.db.Prepare("SELECT func_InsertUser($1, $2, $3)")
	if err != nil {
		return 0, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	// Retrieve the OUT parameter value
	err = stmt.QueryRow(user.Username, user.Email, user.Password).Scan(&userId)
	if err != nil {
		return 0, fmt.Errorf("error executing function: %w", err)
	}

	return userId, nil
}

func (r *adminConsoleRepository) Login(creds models.AdminCreds) (string, string, int, error) {
	var passwordHash string
	var username string
	userId := 0
	stmt, err := r.db.Prepare("Select * From func_getAdminLoginData($1)")
	if err != nil {
		return passwordHash, username, userId, fmt.Errorf("error executing function: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(creds.Email).Scan(&passwordHash, &username, &userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return passwordHash, username, userId, err
		}
		return passwordHash, username, userId, fmt.Errorf("error executing function: %w", err)
	}
	return passwordHash, username, userId, nil
}

func (r *adminConsoleRepository) GetGames() models.GameData {
	var x models.GameData
	return x
}
