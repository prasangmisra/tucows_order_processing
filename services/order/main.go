package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"order/database"
	"order/handlers"
	"order/redisconn"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func main() {

	orderServicePort := os.Getenv("ORDER_SERVICE_PORT")

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer db.Close()

	rdb, err := redisconn.Connect()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	channel := redisconn.Subscribe(rdb, "payment_results")

	go handlePaymentResults(db, channel)

	router := setupRouter(db, rdb)
	fmt.Println("Order Management Service is running on port " + orderServicePort)
	http.ListenAndServe(":"+orderServicePort, router)

}

func handlePaymentResults(db *sql.DB, channel <-chan *redis.Message) {
	ctx := context.Background()
	for msg := range channel {
		var notification map[string]interface{}
		err := json.Unmarshal([]byte(msg.Payload), &notification)
		if err != nil {
			log.Printf("Error unmarshalling payment result notification: %v\n", err)
			continue
		}
		orderID := int(notification["order_id"].(float64))
		status := notification["status"].(string)

		_, err = db.ExecContext(ctx, "UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2", status, orderID)
		if err != nil {
			log.Printf("Error updating order status in order service: %v\n", err)
			continue
		}
	}
}

func setupRouter(db *sql.DB, rdb *redis.Client) *mux.Router {
	router := mux.NewRouter()

	// Load OpenAPI specification
	specPath := filepath.Join("reference", "openapi.yml")
	swagger, err := openapi3.NewLoader().LoadFromFile(specPath)
	if err != nil {
		log.Fatalf("Error loading OpenAPI spec: %v", err)
	}

	// Validate OpenAPI spec
	if err := swagger.Validate(context.Background()); err != nil {
		log.Fatalf("Invalid OpenAPI spec: %v", err)
	}

	// Create router based on OpenAPI spec
	oapiRouter, err := gorillamux.NewRouter(swagger)
	if err != nil {
		log.Fatalf("Error creating OpenAPI router: %v", err)
	}

	router.HandleFunc("/product", handlers.CreateProductHandler(db)).Methods("POST")
	router.HandleFunc("/product/{id}", handlers.GetProductHandler(db)).Methods("GET")
	router.HandleFunc("/order", handlers.CreateOrderHandler(db, rdb)).Methods("POST")
	router.HandleFunc("/order/{id}", handlers.GetOrderHandler(db)).Methods("GET")

	// Add OpenAPI validation middleware
	router.Use(oapiMiddleware(oapiRouter))

	return router
}

func oapiMiddleware(router routers.Router) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error finding route: %v", err), http.StatusBadRequest)
				return
			}

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
			}

			if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
				http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
				return
			}

			// Capture the response status and body
			rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rr, r)

			responseValidationInput := &openapi3filter.ResponseValidationInput{
				RequestValidationInput: requestValidationInput,
				Status:                 rr.statusCode,
				Header:                 w.Header(),
				Body:                   io.NopCloser(bytes.NewReader(rr.body)),
			}

			if err := openapi3filter.ValidateResponse(context.Background(), responseValidationInput); err != nil {
				log.Printf("Response validation failed: %v", err)
			}
		})
	}
}

// responseRecorder is a custom response writer to capture status and body
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.body = append(rr.body, b...)
	return rr.ResponseWriter.Write(b)
}
