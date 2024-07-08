package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

var ctx = context.Background()

const (
	PAYMENT_SUCCESS   = "success"
	PAYMENT_FAILURE   = "failure"
	PAYMENT_THRESHOLD = 1000
)

func main() {
	redisPort := os.Getenv("REDIS_PORT")
	paymentServicePort := os.Getenv("PAYMENT_SERVICE_PORT")

	rdb, err := connectToRedis("redis:" + redisPort)
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v\n", err)
	}

	done := make(chan bool)
	go processPaymentRequests(rdb, done)

	fmt.Println("Payment Processing Service is running on port " + paymentServicePort)
	http.ListenAndServe(":"+paymentServicePort, nil)
}

func connectToRedis(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

func notifyOrderService(orderID uuid.UUID, status string, rdb *redis.Client) {
	notification := map[string]interface{}{
		"order_id": orderID.String(),
		"status":   status,
	}
	notificationJSON, _ := json.Marshal(notification)
	rdb.Publish(ctx, "payment_results", notificationJSON)
	fmt.Printf("Payment result notification sent for order %s with status %s\n", orderID, status)
}

func processPaymentRequests(rdb *redis.Client, done chan bool) {
	subscriber := rdb.Subscribe(ctx, "payment_requests")
	channel := subscriber.Channel()
	for msg := range channel {
		var paymentRequest map[string]interface{}
		err := json.Unmarshal([]byte(msg.Payload), &paymentRequest)
		if err != nil {
			log.Printf("Error unmarshalling payment request: %v\n", err)
			continue
		}

		status := PAYMENT_SUCCESS
		if paymentRequest["amount"].(float64) > PAYMENT_THRESHOLD {
			status = PAYMENT_FAILURE
		}

		orderID, err := uuid.Parse(paymentRequest["order_id"].(string))
		if err != nil {
			log.Printf("Error parsing order ID: %v\n", err)
			continue
		}

		notifyOrderService(orderID, status, rdb)
	}

	done <- true
}
