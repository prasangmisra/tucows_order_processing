package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

var ctx = context.Background()

const (
	DB_USER      = "user"
	DB_PASS      = "password"
	DB_PORT      = "5432"
	DB_DATABASE  = "orderdb"
	REDIS_PORT   = "6379"
	SERVICE_PORT = "8080"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://"+DB_USER+":"+DB_PASS+"@db:"+DB_PORT+"/"+DB_DATABASE+"")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:" + REDIS_PORT,
	})

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v\n", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Order Management Service")
	}).Methods("GET")

	fmt.Println("Order Management Service is running on port " + SERVICE_PORT)
	http.ListenAndServe(":"+SERVICE_PORT, router)
}
