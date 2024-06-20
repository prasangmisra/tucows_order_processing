package models

import "time"

type PaymentRequest struct {
	OrderID    int       `json:"order_id"`
	Amount     float64   `json:"amount"`
	CustomerID int       `json:"customer_id"`
	ProductID  int       `json:"product_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
