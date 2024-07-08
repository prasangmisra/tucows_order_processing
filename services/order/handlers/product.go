package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"order/models"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var validate = validator.New()

func CreateProductHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var productWrite models.ProductWrite
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewDecoder(r.Body).Decode(&productWrite); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request payload"})
			return
		}

		// Validate the productWrite struct
		if err := validate.Struct(productWrite); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
			return
		}

		var productRead models.ProductRead
		err := db.QueryRow("INSERT INTO products (name, price) VALUES ($1, $2) RETURNING id, name, price, created_at, updated_at", productWrite.Name, productWrite.Price).Scan(&productRead.ID, &productRead.Name, &productRead.Price, &productRead.CreatedAt, &productRead.UpdatedAt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to create product: " + err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(productRead)
	}
}

func GetProductHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		w.Header().Set("Content-Type", "application/json")
		id := vars["id"]
		_, err := uuid.Parse(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid product ID"})
			return
		}

		var productRead models.ProductRead
		err = db.QueryRow("SELECT id, name, price, created_at, updated_at FROM products where id = $1", id).Scan(&productRead.ID, &productRead.Name, &productRead.Price, &productRead.CreatedAt, &productRead.UpdatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Product not found"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve product"})
			}
			return
		}

		json.NewEncoder(w).Encode(productRead)
	}
}
