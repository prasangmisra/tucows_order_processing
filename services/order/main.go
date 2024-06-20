package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"order/database"
	"order/handlers"
	"order/redisconn"

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

	router.HandleFunc("/product", handlers.CreateProductHandler(db)).Methods("POST")
	router.HandleFunc("/product/{id}", handlers.GetProductHandler(db)).Methods("GET")
	router.HandleFunc("/order", handlers.CreateOrderHandler(db, rdb)).Methods("POST")
	router.HandleFunc("/order/{id}", handlers.GetOrderHandler(db)).Methods("GET")

	return router
}
