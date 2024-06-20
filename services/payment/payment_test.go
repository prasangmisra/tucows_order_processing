package main

import (
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConnectToRedis(t *testing.T) {
	// Start a mini Redis server
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	// Connect to the mini Redis server
	rdb, err := connectToRedis(mr.Addr())
	assert.NoError(t, err)
	assert.NotNil(t, rdb)
}

func TestNotifyOrderService(t *testing.T) {
	// Start a mini Redis server
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	// Connect to the mini Redis server
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Subscribe to the channel
	subscriber := rdb.Subscribe(ctx, "payment_results")
	defer subscriber.Close()

	// Wait for subscription to be established
	_, err = subscriber.Receive(ctx)
	assert.NoError(t, err)

	// Send a notification
	orderID := uuid.New()
	status := PAYMENT_SUCCESS
	notifyOrderService(orderID, status, rdb)

	// Read the message from the channel
	msg, err := subscriber.ReceiveMessage(ctx)
	assert.NoError(t, err)

	var notification map[string]interface{}
	err = json.Unmarshal([]byte(msg.Payload), &notification)
	assert.NoError(t, err)
	assert.Equal(t, orderID.String(), notification["order_id"].(string))
	assert.Equal(t, status, notification["status"].(string))
}

func TestProcessPaymentRequests(t *testing.T) {
	// Start a mini Redis server
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	// Connect to the mini Redis server
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Use a channel to signal when the processing is done
	done := make(chan bool)

	// Start processing payment requests
	go processPaymentRequests(rdb, done)

	// Subscribe to the payment results channel
	subscriber := rdb.Subscribe(ctx, "payment_results")
	defer subscriber.Close()

	// Wait for subscription to be established
	_, err = subscriber.Receive(ctx)
	assert.NoError(t, err)

	// Send a payment request
	orderID := uuid.New()
	paymentRequest := map[string]interface{}{
		"order_id": orderID.String(),
		"amount":   500,
	}
	payload, err := json.Marshal(paymentRequest)
	assert.NoError(t, err)
	rdb.Publish(ctx, "payment_requests", payload)

	// Read the message from the channel
	msg, err := subscriber.ReceiveMessage(ctx)
	assert.NoError(t, err)

	var notification map[string]interface{}
	err = json.Unmarshal([]byte(msg.Payload), &notification)
	assert.NoError(t, err)
	assert.Equal(t, orderID.String(), notification["order_id"].(string))
	assert.Equal(t, PAYMENT_SUCCESS, notification["status"].(string))
}
