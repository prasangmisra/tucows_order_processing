package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"order/handlers"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib" // Import the pgx driver
)

var ctx = context.Background()

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	redisPort := os.Getenv("REDIS_PORT")
	orderServicePort := os.Getenv("ORDER_SERVICE_PORT")

	connStr := "postgres://" + dbUser + ":" + dbPass + "@db:" + dbPort + "/" + dbName
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:" + redisPort,
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

	fmt.Println("Order Management Service is running on port " + orderServicePort)
	http.ListenAndServe(":"+orderServicePort, router)
}
