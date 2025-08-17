package services

import (
	"GameWala-Arcade/models"
	"GameWala-Arcade/repositories"
)

type HandlePaymentService interface {
	SaveOrderDetails(models.PaymentStatus) error
}

type handlePaymentService struct {
	handlePaymentRepository repositories.HandlePaymentRepository
}

func NewHandlePaymentService(handlePaymentRepository repositories.HandlePaymentRepository) *handlePaymentService {
	return &handlePaymentService{handlePaymentRepository: handlePaymentRepository}
}

func (s *handlePaymentService) SaveOrderDetails(details models.PaymentStatus) error {
	err := s.handlePaymentRepository.SaveOrderDetails(details)
	return err
}
