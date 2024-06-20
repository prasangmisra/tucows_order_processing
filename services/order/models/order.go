package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	Pending   OrderStatus = "pending"
	Completed OrderStatus = "completed"
	Failed    OrderStatus = "failed"
)

type OrderWrite struct {
	CustomerID uuid.UUID `json:"customer_id" validate:"required"`
	ProductID  uuid.UUID `json:"product_id" validate:"required"`
	Amount     float64   `json:"amount" validate:"required"`
}

type OrderRead struct {
	ID         uuid.UUID   `json:"id"`
	CustomerID uuid.UUID   `json:"customer_id"`
	ProductID  uuid.UUID   `json:"product_id"`
	Status     OrderStatus `json:"status"`
	Amount     float64     `json:"amount"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
