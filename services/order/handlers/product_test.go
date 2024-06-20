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
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestCreateProductHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := CreateProductHandler(db)

	t.Run("successful product creation", func(t *testing.T) {
		productWrite := models.ProductWrite{Name: "New Product", Price: 99.99}
		productRead := models.ProductRead{ID: 1, Name: "New Product", Price: 99.99}

		mock.ExpectQuery("INSERT INTO products").
			WithArgs(productWrite.Name, productWrite.Price).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).AddRow(productRead.ID, productRead.Name, productRead.Price))

		body, _ := json.Marshal(productWrite)
		req, err := http.NewRequest("POST", "/product", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result models.ProductRead
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, productRead, result)
	})

	t.Run("invalid request payload", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/product", bytes.NewBuffer([]byte(`{"name": ""}`)))
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

		assert.Equal(t, models.ErrorResponse{Error: "Key: 'ProductWrite.Name' Error:Field validation for 'Name' failed on the 'required' tag\nKey: 'ProductWrite.Price' Error:Field validation for 'Price' failed on the 'required' tag"}, result)
	})
}

func TestGetProductHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := GetProductHandler(db)

	t.Run("successful product retrieval", func(t *testing.T) {
		productRead := models.ProductRead{ID: 1, Name: "Sample Product", Price: 99.99}

		mock.ExpectQuery("SELECT id, name, price FROM products where id = \\$1").
			WithArgs(productRead.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).AddRow(productRead.ID, productRead.Name, productRead.Price))

		req, err := http.NewRequest("GET", "/product/"+strconv.Itoa(productRead.ID), nil)
		if err != nil {
			t.Fatal(err)
		}

		req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(productRead.ID)})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.ProductRead
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, productRead, result)
	})

	t.Run("product not found", func(t *testing.T) {
		productID := 2

		mock.ExpectQuery("SELECT id, name, price FROM products where id = \\$1").
			WithArgs(productID).
			WillReturnError(sql.ErrNoRows)

		req, err := http.NewRequest("GET", "/product/"+strconv.Itoa(productID), nil)
		if err != nil {
			t.Fatal(err)
		}

		req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(productID)})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		var result models.ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, models.ErrorResponse{Error: "Product not found"}, result)
	})

	t.Run("invalid product ID", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/product/invalid", nil)
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

		assert.Equal(t, models.ErrorResponse{Error: "Invalid product ID"}, result)
	})
}
