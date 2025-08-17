package models

type PaymentStatus struct {
	OrderCreationId   string
	RazorpayPaymentId string
	RazorpayOrderId   string
	RazorPaySignature string
}
