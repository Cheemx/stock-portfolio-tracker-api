# Stock Portfolio Tracker API

A stock tracking and portfolio management API built with **Golang**, **PostgreSQL**, **Redis**, and **Docker**.
The project focuses on clean modular code, DRY principles, and efficient backend SQL queries while exploring **background workers** using **Redis Streams** (kind of Message queue), authentication, and scalable API design.

---

## Requirements

* [Golang](https://go.dev/)
* [PostgreSQL](https://www.postgresql.org/)
* [Redis](https://redis.io/)
* [Docker](https://www.docker.com/)

You can use locally installed PostgreSQL and Redis, or run them via Docker images.
I'd prefer using Docker personally and copy the env variables from `.env.sample` and create your own .env to run it.

---

## Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/Cheemx/stock-portfolio-tracker-api.git
   cd stock-portfolio-tracker-api
   go install github.com/pressly/goose/v3/cmd/goose@latest 
   # To run migrations installing goose is necessary.
   ```

2. **Run database migrations** (using Makefile and Goose)

   ```bash
   make migrationUp
   ```

   * To rollback migrations:

     ```bash
     make migrationDown
     ```
   * To connect to `stonk` database in CLI:

     ```bash
     make databaseDikha
     ```

3. **Start services with Docker Compose**

   ```bash
   docker compose up -d
   ```

4. **Run the application**

   ```bash
   go run .
   ```

5. The application will be live at:

   ```
   http://localhost:8080
   ```

---


## Features

### Authentication & Users

* **POST /api/users** – Create a new user
* **POST /api/login** – Login and receive a JWT token
* Passwords secured with **bcrypt**
* JWT-based authentication and authorization (required for all routes except user creation and login)

### Transactions

* **POST /api/transactions** – Record BUY/SELL operations
* Stores both **transaction history** and updates **user holdings** with average price and total invested

### Portfolio & Holdings

* **GET /api/portfolio** – Fetch portfolio summary (total invested, current value, PnL, holdings count)
* **GET /api/holdings** – Get all holdings with live prices, evaluations, and PnL

### Transaction History

* **GET /api/transactions** – Retrieve all user transactions

### Stocks

* **GET /api/stocks/search?q=apple** - Search for a stock by stock symbol or company name
* **GET /api/stocks** - Fetch recent 10 stocks

### Background Worker

* Redis Streams worker fetches stock data from Yahoo API every 30 seconds
* Updated prices are dumped into the **stocks** table in PostgreSQL

---

## Example API Workflows

### User Creation

**Request**

```json
{
    "name": "Cheems",
    "email": "cheems@gmail.com",
    "password": "cheems"
}
```

**Response**

```json
{
    "id": "290fc0aa-aaad-45c7-a6a4-88aa4ab4291f",
    "name": "Cheems",
    "created_at": "2025-09-20T13:36:57.573341Z",
    "email": "cheems@gmail.com"
}
```

---

### Login

**Request**

```json
{
    "email": "cheems@gmail.com",
    "password": "cheems"
}
```

**Response**

```json
{
    "id": "290fc0aa-aaad-45c7-a6a4-88aa4ab4291f",
    "created_at": "2025-09-20T13:36:57.573341Z",
    "email": "cheems@gmail.com",
    "name": "Cheems",
    "token": "<JWT Token>" //store this token for Authorization header
}
```

---

### Transaction (BUY)

**Request**

```json
{
    "stock_symbol": "MSFT",
    "type": "BUY",
    "quantity": 2
}
```

**Response**

```json
{
    "Holding": {
        "id": "af589d82-6f77-4210-a8b9-1fdb0974e7b8",
        "user_id": "290fc0aa-aaad-45c7-a6a4-88aa4ab4291f",
        "stock_symbol": "MSFT",
        "quantity": 2,
        "average_price": 517.93,
        "created_at": "2025-09-20T13:38:18.676128Z",
        "updated_at": "2025-09-20T13:38:18.676128Z",
        "total_invested": 1035.86
    },
    "Transaction": {
        "id": "e33f93a8-8dca-4a75-8057-b81a06e79a47",
        "user_id": "290fc0aa-aaad-45c7-a6a4-88aa4ab4291f",
        "stock_symbol": "MSFT",
        "type": "BUY",
        "quantity": 2,
        "price": 517.93,
        "total_amount": 1035.86,
        "created_at": "2025-09-20T13:38:18.667612Z"
    }
}
```

---

### Portfolio Summary

**Response**

```json
{
    "total_invested": 1035.86,
    "current_value": 1035.86,
    "pnl": 0,
    "pnl_percentage": 0,
    "holdings_count": 1
}
```

---

### Holdings

**Response**

```json
[
    {
        "stock_symbol": "MSFT",
        "company_name": "Microsoft Corporation",
        "quantity": 2,
        "average_price": 517.93,
        "curr_price": 517.93,
        "curr_evaluation": 1035.86,
        "pnl": 0,
        "pnl_percentage": 0,
        "total_invested": 1035.86
    }
]
```
---

### Search and Get Stocks

**Response**

```json
[
    {
        "symbol": "MSFT",
        "company_name": "Microsoft Corporation",
        "current_price": 517.93,
        "previous_close": 415.00,
        "updated_at": "2025-09-20T13:38:18.667612Z"
    },
    // ...
]
```

---

## Authentication

* All routes except `/api/users` and `/api/login` require an **Authorization** header:

  ```
  Authorization: Bearer <JWT Token>
  ```

---

## Technologies Used

* **Golang** – REST API and background worker
* **PostgreSQL** – Relational database for users, transactions, stocks and holdings
* **Redis** – Streams used for background stock price ingestion
* **Docker** – Containerized environment
* **JWT & Bcrypt** – Authentication and authorization

---

## Planned Improvements Improving Infrastructure.

* Rate limiter for API security
* Nginx for load balancing
* WebSocket support for live price updates