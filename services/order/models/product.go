package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductWrite represents a product for creating or updating
type ProductWrite struct {
	Name  string  `json:"name" validate:"required"`
	Price float64 `json:"price" validate:"required"`
}

// ProductRead represents a product for reading
type ProductRead struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
