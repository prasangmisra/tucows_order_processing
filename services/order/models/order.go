package models

import "time"

type OrderWrite struct {
	CustomerID int     `json:"customer_id" validate:"required"`
	ProductID  int     `json:"product_id" validate:"required"`
	Amount     float64 `json:"amount" validate:"required"`
}

type OrderRead struct {
	ID         int       `json:"id"`
	CustomerID int       `json:"customer_id"`
	ProductID  int       `json:"product_id"`
	Status     string    `json:"status"`
	Amount     float64   `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
