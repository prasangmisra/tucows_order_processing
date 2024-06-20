package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"order/models"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
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
		orderWrite := models.OrderWrite{CustomerID: 1, ProductID: 1, Amount: 100.0}
		orderRead := models.OrderRead{ID: 1, CustomerID: 1, ProductID: 1, Status: "pending", Amount: 100.0}

		mock.ExpectQuery("INSERT INTO orders").
			WithArgs(orderWrite.CustomerID, orderWrite.ProductID, orderWrite.Amount).
			WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "product_id", "status", "amount"}).
				AddRow(orderRead.ID, orderRead.CustomerID, orderRead.ProductID, orderRead.Status, orderRead.Amount))

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

		assert.Equal(t, orderRead, result)
	})

	t.Run("invalid request payload", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/order", bytes.NewBuffer([]byte(`{"customer_id": 1}`)))
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

		assert.Equal(t, models.ErrorResponse{Error: "Key: 'OrderWrite.ProductID' Error:Field validation for 'ProductID' failed on the 'required' tag\nKey: 'OrderWrite.Amount' Error:Field validation for 'Amount' failed on the 'required' tag"}, result)
	})

	t.Run("database error on order creation", func(t *testing.T) {
		orderWrite := models.OrderWrite{CustomerID: 1, ProductID: 1, Amount: 100.0}

		mock.ExpectQuery("INSERT INTO orders").
			WithArgs(orderWrite.CustomerID, orderWrite.ProductID, orderWrite.Amount).
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

		assert.Equal(t, models.ErrorResponse{Error: "Failed to create order: sql: connection is already closed"}, result)
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
		orderRead := models.OrderRead{ID: 1, CustomerID: 1, ProductID: 1, Status: "pending", Amount: 100.0}

		mock.ExpectQuery("SELECT id, customer_id, product_id, status, amount FROM orders WHERE id = \\$1").
			WithArgs(orderRead.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "product_id", "status", "amount"}).
				AddRow(orderRead.ID, orderRead.CustomerID, orderRead.ProductID, orderRead.Status, orderRead.Amount))

		req, err := http.NewRequest("GET", "/order/"+strconv.Itoa(orderRead.ID), nil)
		if err != nil {
			t.Fatal(err)
		}

		req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(orderRead.ID)})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.OrderRead
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, orderRead, result)
	})

	t.Run("order not found", func(t *testing.T) {
		orderID := 2

		mock.ExpectQuery("SELECT id, customer_id, product_id, status, amount FROM orders WHERE id = \\$1").
			WithArgs(orderID).
			WillReturnError(sql.ErrNoRows)

		req, err := http.NewRequest("GET", "/order/"+strconv.Itoa(orderID), nil)
		if err != nil {
			t.Fatal(err)
		}

		req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(orderID)})
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
