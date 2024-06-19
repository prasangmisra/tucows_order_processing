package main

import (
	"fmt"
	"net/http"
)

const (
	SERVICE_PORT = "8081"
)

func main() {
	http.HandleFunc("/payment", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Payment Processing Service")
	})

	fmt.Println("Payment Processing Service is running on port " + SERVICE_PORT)
	http.ListenAndServe(":"+SERVICE_PORT, nil)
}
