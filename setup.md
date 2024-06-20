# Setup Guide for Tucows Order Processing

## Prerequisites

Ensure you have the following installed:
- Docker
- Docker Compose

## Project Structure

tucows_order_processing/
├── docker-compose.yml
├── init.sql
├── services/
│ ├── order/
│ │ ├── Dockerfile
│ │ └── main.go
│ ├── payment/
│ │ ├── Dockerfile
│ │ └── main.go
└── setup.md

## Running the Application

1. Clone the repository:
   ```bash
   git clone https://github.com/prasangmisra/tucows_order_processing.git
   ```

2. cd tucows-interview-exercise

3. Ensure `init.sql` contains the necessary SQL commands to initialize the database.

4. ```docker-compose up --build```

5. Ensure all services are running by `docker ps`.

5. The following services will be available:

 - Order Management Service: `http://localhost:8080`

 -  Payment Processing Service: `http://localhost:8081`

 -  Postgres Database: `localhost:5432`

 - Redis: `localhost:6379`

 ## Endpoints
 ### Order Service Endpoints
 1. POST /order: Creates a new order
 - Request: `curl -X POST localhost:8080/order -H "Content-Type: application/json" -d '{"customer_id": "<uuid>", "product_id": "<uuid>", "amount": <amount>}'`
 - Response: `{"id": "<uuid>", "customer_id":"<uuid>", "product_id":"<uuid>", "status":"<status>", "amount": <amount>, "created_at":"<created_at>", "updated_at":"<updated_at>"}`

  2. GET /order/id: Fetches order by ID
 - Request: `curl -X GET http://localhost:8080/order/<uuid>`
 - Response: `{"id": "<uuid>", "customer_id":"<uuid>", "product_id":"<uuid>", "status":"<status>", "amount": <amount>, "created_at":"<created_at>", "updated_at":"<updated_at>"}`

