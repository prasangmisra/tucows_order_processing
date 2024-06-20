package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"order/handlers"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib" // Import the pgx driver
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
	connStr := "postgres://" + DB_USER + ":" + DB_PASS + "@db:" + DB_PORT + "/" + DB_DATABASE
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:" + REDIS_PORT,
	})

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v\n", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/product", handlers.CreateProductHandler(db)).Methods("POST")
	router.HandleFunc("/product/{id}", handlers.GetProductHandler(db)).Methods("GET")
	router.HandleFunc("/order", handlers.CreateOrderHandler(db, rdb)).Methods("POST")
	router.HandleFunc("/order/{id}", handlers.GetOrderHandler(db)).Methods("GET")

	fmt.Println("Order Management Service is running on port " + SERVICE_PORT)
	http.ListenAndServe(":"+SERVICE_PORT, router)
}
