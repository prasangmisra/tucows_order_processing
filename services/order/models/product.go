package models

// ProductWrite represents a product for creating or updating
type ProductWrite struct {
	Name  string  `json:"name" validate:"required"`
	Price float64 `json:"price" validate:"required"`
}

// ProductRead represents a product for reading
type ProductRead struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
