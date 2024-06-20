package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentRequest struct {
	OrderID    uuid.UUID `json:"order_id"`
	Amount     float64   `json:"amount"`
	CustomerID uuid.UUID `json:"customer_id"`
	ProductID  uuid.UUID `json:"product_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
