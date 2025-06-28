package repositories

import (
	"database/sql"
)

type HandlePaymentRepository interface {
}

type handlePaymentRepository struct {
	db *sql.DB
}

func NewHandlePaymentReposiory(db *sql.DB) *handlePaymentRepository {
	return &handlePaymentRepository{db: db}
}
