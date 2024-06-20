package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"order/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrderHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	handler := CreateOrderHandler(db, rdb)

	t.Run("successful order creation", func(t *testing.T) {
		customerID := uuid.New()
		productID := uuid.New()
		orderID := uuid.New()
		orderWrite := models.OrderWrite{CustomerID: customerID, ProductID: productID, Amount: 100.0}
		orderRead := models.OrderRead{
			ID:         orderID,
			CustomerID: customerID,
			ProductID:  productID,
			Status:     models.Pending,
			Amount:     100.0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mock.ExpectQuery("INSERT INTO orders").
			WithArgs(orderWrite.CustomerID, orderWrite.ProductID, models.Pending, orderWrite.Amount).
			WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "product_id", "status", "amount", "created_at", "updated_at"}).
				AddRow(orderRead.ID, orderRead.CustomerID, orderRead.ProductID, orderRead.Status, orderRead.Amount, orderRead.CreatedAt, orderRead.UpdatedAt))

		body, _ := json.Marshal(orderWrite)
		req, err := http.NewRequest("POST", "/order", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result models.OrderRead
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, orderRead.CustomerID, result.CustomerID)
		assert.Equal(t, orderRead.ProductID, result.ProductID)
		assert.Equal(t, orderRead.Status, result.Status)
		assert.Equal(t, orderRead.Amount, result.Amount)
		assert.WithinDuration(t, orderRead.CreatedAt, result.CreatedAt, time.Second)
		assert.WithinDuration(t, orderRead.UpdatedAt, result.UpdatedAt, time.Second)
	})

	t.Run("invalid request payload", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/order", bytes.NewBuffer([]byte(`{"customer_id": "invalid-uuid"}`)))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var result models.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, models.ErrorResponse{Error: "Invalid request payload"}, result)
	})

	t.Run("database error on order creation", func(t *testing.T) {
		customerID := uuid.New()
		productID := uuid.New()
		orderWrite := models.OrderWrite{CustomerID: customerID, ProductID: productID, Amount: 100.0}

		mock.ExpectQuery("INSERT INTO orders").
			WithArgs(orderWrite.CustomerID, orderWrite.ProductID, models.Pending, orderWrite.Amount).
			WillReturnError(sql.ErrConnDone)

		body, _ := json.Marshal(orderWrite)
		req, err := http.NewRequest("POST", "/order", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var result models.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, models.ErrorResponse{Error: "Failed to create order: " + sql.ErrConnDone.Error()}, result)
	})
}

func TestGetOrderHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := GetOrderHandler(db)

	t.Run("successful order retrieval", func(t *testing.T) {
		customerID := uuid.New()
		productID := uuid.New()
		orderID := uuid.New()
		orderRead := models.OrderRead{
			ID:         orderID,
			CustomerID: customerID,
			ProductID:  productID,
			Status:     models.Pending,
			Amount:     100.0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mock.ExpectQuery("SELECT id, customer_id, product_id, status, amount, created_at, updated_at FROM orders WHERE id = \\$1").
			WithArgs(orderRead.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "product_id", "status", "amount", "created_at", "updated_at"}).
				AddRow(orderRead.ID, orderRead.CustomerID, orderRead.ProductID, orderRead.Status, orderRead.Amount, orderRead.CreatedAt, orderRead.UpdatedAt))

		req, err := http.NewRequest("GET", "/order/"+orderRead.ID.String(), nil)
		if err != nil {
			t.Fatal(err)
		}

		req = mux.SetURLVars(req, map[string]string{"id": orderRead.ID.String()})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.OrderRead
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, orderRead.CustomerID, result.CustomerID)
		assert.Equal(t, orderRead.ProductID, result.ProductID)
		assert.Equal(t, orderRead.Status, result.Status)
		assert.Equal(t, orderRead.Amount, result.Amount)
		assert.WithinDuration(t, orderRead.CreatedAt, result.CreatedAt, time.Second)
		assert.WithinDuration(t, orderRead.UpdatedAt, result.UpdatedAt, time.Second)
	})

	t.Run("order not found", func(t *testing.T) {
		orderID := uuid.New().String()

		mock.ExpectQuery("SELECT id, customer_id, product_id, status, amount, created_at, updated_at FROM orders WHERE id = \\$1").
			WithArgs(orderID).
			WillReturnError(sql.ErrNoRows)

		req, err := http.NewRequest("GET", "/order/"+orderID, nil)
		if err != nil {
			t.Fatal(err)
		}

		req = mux.SetURLVars(req, map[string]string{"id": orderID})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		var result models.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, models.ErrorResponse{Error: "Order not found"}, result)
	})

	t.Run("invalid order ID", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/order/invalid", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var result models.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, models.ErrorResponse{Error: "Invalid order ID"}, result)
	})
}
