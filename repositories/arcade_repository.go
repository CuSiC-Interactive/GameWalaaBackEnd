package repositories

import (
	"GameWala-Arcade/utils"
	"database/sql"
)

type ArcadeRepository interface {
	ValidateArcade(arcadeId string) (bool, error)
}

type arcadeRepository struct {
	db *sql.DB
}

func NewArcadeRepository(db *sql.DB) *arcadeRepository {
	return &arcadeRepository{db: db}
}

func (r *arcadeRepository) ValidateArcade(arcadeId string) (bool, error) {
	// Implement the logic to validate the arcade ID against the database.
	// For now, we will return true for any non-empty arcade ID.
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM "Arcade" WHERE id = $1`, arcadeId).Scan(&count)
	if err != nil {
		utils.LogError("Getting some error while executing statement on DB %v", err)
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}
