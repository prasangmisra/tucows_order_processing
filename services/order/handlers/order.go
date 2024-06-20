package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"order/models"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var ctx = context.Background()

func CreateOrderHandler(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orderWrite models.OrderWrite
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewDecoder(r.Body).Decode(&orderWrite); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request payload"})
			return
		}

		// Validate the orderWrite struct
		if err := validate.Struct(orderWrite); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
			return
		}

		var orderRead models.OrderRead
		err := db.QueryRow(
			"INSERT INTO orders (customer_id, product_id, status, amount, created_at, updated_at) VALUES ($1, $2, 'pending', $3, NOW(), NOW()) RETURNING id, customer_id, product_id, status, amount, created_at, updated_at",
			orderWrite.CustomerID, orderWrite.ProductID, orderWrite.Amount).Scan(
			&orderRead.ID, &orderRead.CustomerID, &orderRead.ProductID, &orderRead.Status, &orderRead.Amount, &orderRead.CreatedAt, &orderRead.UpdatedAt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to create order: " + err.Error()})
			return
		}

		// Send payment request to payment service
		paymentRequest := map[string]interface{}{
			"order_id":    orderRead.ID,
			"amount":      orderRead.Amount,
			"customer_id": orderRead.CustomerID,
			"product_id":  orderRead.ProductID,
		}
		paymentRequestJSON, _ := json.Marshal(paymentRequest)
		rdb.Publish(ctx, "payment_requests", paymentRequestJSON)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(orderRead)
	}
}

func GetOrderHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		w.Header().Set("Content-Type", "application/json")
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid order ID"})
			return
		}

		var orderRead models.OrderRead
		err = db.QueryRow(
			"SELECT id, customer_id, product_id, status, amount, created_at, updated_at FROM orders WHERE id = $1",
			id).Scan(
			&orderRead.ID, &orderRead.CustomerID, &orderRead.ProductID, &orderRead.Status, &orderRead.Amount, &orderRead.CreatedAt, &orderRead.UpdatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Order not found"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve order"})
			}
			return
		}

		json.NewEncoder(w).Encode(orderRead)
	}
}
