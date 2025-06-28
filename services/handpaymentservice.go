package services

import (
	"GameWala-Arcade/repositories"
)

type HandlePaymentService interface {
}

type handlePaymentService struct {
	handlePaymentRepository repositories.HandlePaymentRepository
}

func NewHandlePaymentService(handlePaymentRepository repositories.HandlePaymentRepository) *handlePaymentService {
	return &handlePaymentService{handlePaymentRepository: handlePaymentRepository}
}
