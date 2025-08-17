package repositories

import (
	"GameWala-Arcade/models"
	"GameWala-Arcade/utils"
	"database/sql"
	"fmt"
)

type HandlePaymentRepository interface {
	SaveOrderDetails(models.PaymentStatus) error
}

type handlePaymentRepository struct {
	db *sql.DB
}

func NewHandlePaymentReposiory(db *sql.DB) *handlePaymentRepository {
	return &handlePaymentRepository{db: db}
}

func (r *handlePaymentRepository) SaveOrderDetails(details models.PaymentStatus) error {
	utils.LogInfo("Saving payment status for payment ID %s", details.RazorpayPaymentId)

	// Prepare the call to the stored procedure
	stmt, err := r.db.Prepare("SELECT func_InsertPaymentStatus($1, $2, $3, $4)")
	if err != nil {
		utils.LogError("Failed to prepare save payment status statement: %v", err)
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(details.OrderCreationId, details.RazorpayPaymentId,
		details.RazorpayOrderId, details.RazorpaySignature)
	if err != nil {
		utils.LogError("Failed to execute payment status for payment ID %s: %v", details.RazorpayPaymentId, err)
		return fmt.Errorf("error executing function: %w", err)
	}

	utils.LogInfo("Successfully saved payment status for order ID %s", details.OrderCreationId)
	return nil
}
