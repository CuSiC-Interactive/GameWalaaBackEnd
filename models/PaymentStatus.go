package models

type PaymentStatus struct {
	OrderCreationId   string
	RazorpayPaymentId string
	RazorpayOrderId   string
	RazorpaySignature string
}
