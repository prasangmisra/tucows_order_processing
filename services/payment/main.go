package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"payment/models"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
)

var ctx = context.Background()

const (
	PAYMENT_SUCCESS = "success"
	PAYMENT_FAILURE = "failure"
)

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	redisPort := os.Getenv("REDIS_PORT")
	paymentServicePort := os.Getenv("PAYMENT_SERVICE_PORT")

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@db:%s/%s", dbUser, dbPass, dbPort, dbName))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:" + redisPort,
	})

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v\n", err)
	}

	subscriber := rdb.Subscribe(ctx, "payment_requests")
	channel := subscriber.Channel()

	go func() {
		for msg := range channel {
			var paymentRequest models.PaymentRequest
			err := json.Unmarshal([]byte(msg.Payload), &paymentRequest)
			if err != nil {
				log.Printf("Error unmarshalling payment request: %v\n", err)
				continue
			}

			status := PAYMENT_SUCCESS
			if paymentRequest.Amount > 1000 {
				status = PAYMENT_FAILURE
			}

			// Notify order service of payment result
			notifyOrderService(paymentRequest.OrderID, status, rdb)
		}
	}()

	fmt.Println("Payment Processing Service is running on port " + paymentServicePort)
	http.ListenAndServe(":"+paymentServicePort, nil)
}

func notifyOrderService(orderID int, status string, rdb *redis.Client) {
	notification := map[string]interface{}{
		"order_id": orderID,
		"status":   status,
	}
	notificationJSON, _ := json.Marshal(notification)
	rdb.Publish(ctx, "payment_results", notificationJSON)
	fmt.Printf("Payment result notification sent for order %d with status %s\n", orderID, status)
}
