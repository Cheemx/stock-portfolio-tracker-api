# Stock Portfolio Tracker API

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-336791?style=flat&logo=postgresql)](https://postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?style=flat&logo=redis)](https://redis.io/)

## Update
After carefully reconsidering the WebSocket implementation, it was overkill since there was no bidirectional messaging requirement. Following a friend's suggestion, re-engineered the real-time communication with **Server Sent Events (SSE)**.

Now it runs on `/api/events` and continuously serves stock data that the user actually holds. If the user holds no stocks, nothing will be served to them making it more efficient and purposeful.

## Key Features

A RESTful API built with **Go** that provides real-time stock portfolio tracking with advanced features including WebSocket connections, Redis caching, rate limiting, and background processing.

- **Real-time Stock Data Integration** - Live market data via Yahoo Finance API
- **Server Sent Events (SSE)** - Real-time portfolio updates to connected clients
- **Caching Strategy** - Redis caching with Configured TTL management and Invalidations
- **Rate Limiting** - Sophisticated request throttling with action-based limits
- **Background Processing** - Asynchronous data fetching and processing
- **RESTful Architecture** - Clean API design and Modular Code structuring
- **Containerization** - Background services used with Docker Containers

## Technical Architecture:

### Tech Stack
- **Backend**: Go 1.25 with gin-gonic
- **Database**: PostgreSQL with Migration management
- **Message Broker**: Redis Streams and SSE for (almost) real-time communication
- **Containerization**: Docker and Docker Compose
- **Authentication**: JWT tokens for Authorization
- **API Integration**: Yahoo Finance for market data

### Performance Optimization
- **Multi-level Caching**: 30s TTL for stocks, 5min for holdings and portfolios
- **Background Workers**: Non-Blocking Data processing
- **Rate Limiting**: Prevents abuse and ensures fair usage
- **Efficient Queries**: Optimized  SQL queries with type-safe Go generated code using sqlc

## Setup

### Prerequisites
- [Golang](https://go.dev/)
- [Docker](https://www.docker.com/)

### Installation

```bash
# Clone the repository
git clone https://github.com/Cheemx/stock-portfolio-tracker-api.git
cd stock-portfolio-tracker-api
go install github.com/pressly/goose/v3/cmd/goose@latest 
# To run migrations installing goose is necessary.

# Start infrastructure services
docker compose up -d

# Copy environment variables
cp .env.sample .env # Populate them with your values
# Mind changing PostgreSQL port in Docker-Cmopose.yml mine is 5433

# Run database migrations
make migrationUp

# Connect to database (optional)
make databaseDikha

# Start the application
go run .
```

The API will be available at `http://localhost:8080`

## API Endpoints and Structure

### Authentication Endpoints

#### Register User
```json
POST /api/users
Content-Type: application/json

{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword"
}
```

**Response:**
```json
{
    "id": "290fc0aa-aaad-45c7-a6a4-88aa4ab4291f",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2025-09-20T13:36:57.573341Z"
}
```

#### Login
```json
POST /api/login
Content-Type: application/json

{
    "email": "john@example.com",
    "password": "securepassword"
}
```

**Response:**
```json
{
    "id": "290fc0aa-aaad-45c7-a6a4-88aa4ab4291f",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2025-09-20T13:36:57.573341Z",
    "token": "<JWT_TOKEN>" //store this token for Authorization header
}
```

### Trading Endpoints

#### Execute Transaction
```json
POST /api/transactions
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
    "stock_symbol": "AAPL",
    "type": "BUY", // Make type 'SELL' for selling
    "quantity": 10
}
```

**Response:**
```json
{
    "Holding": {
        "id": "af589d82-6f77-4210-a8b9-1fdb0974e7b8",
        "user_id": "290fc0aa-aaad-45c7-a6a4-88aa4ab4291f",
        "stock_symbol": "AAPL",
        "quantity": 10,
        "average_price": 150.25,
        "total_invested": 1502.50,
        "created_at": "2025-09-20T13:38:18.676128Z",
        "updated_at": "2025-09-20T13:38:18.676128Z"
    },
    "Transaction": {
        "id": "e33f93a8-8dca-4a75-8057-b81a06e79a47",
        "user_id": "290fc0aa-aaad-45c7-a6a4-88aa4ab4291f",
        "stock_symbol": "AAPL",
        "type": "BUY",
        "quantity": 10,
        "price": 150.25,
        "total_amount": 1502.50,
        "created_at": "2025-09-20T13:38:18.667612Z"
    }
}
```

### Portfolio Management

#### Get Portfolio Summary
```json
GET /api/portfolio
Authorization: Bearer <JWT_TOKEN>
```

**Response:**
```json
{
    "total_invested": 1502.50,
    "current_value": 1587.30,
    "pnl": 84.80,
    "pnl_percentage": 5.64,
    "holdings_count": 3
}
```

#### Get Holdings
```json
GET /api/holdings
Authorization: Bearer <JWT_TOKEN>
```

**Response:**
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
    },
    // ...
]
```

### Market Data

#### Get Recent Stocks
```json
GET /api/stocks
```

**Response:**
```json
[
    {
        "symbol": "^NSEI",
        "company_name": "NIFTY 50",
        "current_price": 25169.5,
        "previous_close": {
            "Float64": 25202.35,
            "Valid": true
        },
        "updated_at": "2025-09-23T16:35:07.262671Z"
    },
    {
        "symbol": "HDFCBANK.NS",
        "company_name": "HDFC Bank Limited",
        "current_price": 957.2,
        "previous_close": {
            "Float64": 964.2,
            "Valid": true
        },
        "updated_at": "2025-09-23T16:35:07.158403Z"
    },
    {
        "symbol": "TCS.NS",
        "company_name": "Tata Consultancy Services Limited",
        "current_price": 3062.4,
        "previous_close": {
            "Float64": 3073.8,
            "Valid": true
        },
        "updated_at": "2025-09-23T16:35:07.027392Z"
    }
    // ...
]
```

#### Search Stocks
```http
GET /api/stocks/search?q=apple
```

**Response:**
```json
[
    {
        "symbol": "AAPL",
        "company_name": "Apple Inc.",
        "current_price": 255.7,
        "previous_close": {
            "Float64": 256.08,
            "Valid": true
        },
        "updated_at": "2025-09-23T16:35:36.564495Z"
    }
]
```

### Real-time Updates

#### Server Sent Events (SSE)
Continuously provides real-time stock updates for user's holdings in JSON format.

```http
GET /api/events
Authorization: Bearer <JWT_TOKEN>
Accept: text/event-stream
```

**JavaScript Client Example:**
```javascript
const eventSource = new EventSource('/api/events', {
    headers: {
        'Authorization': 'Bearer ' + token
    }
});

eventSource.onmessage = function(event) {
    const stockData = JSON.parse(event.data);
    console.log('Real-time stock update:', stockData);
};
```

**Event Stream Response Format:**
```
data: {"symbol":"AAPL","company_name":"Apple Inc.","current_price":255.7,"previous_close":256.08,"updated_at":"2025-09-23T16:35:36Z"}

data: {"symbol":"MSFT","company_name":"Microsoft Corporation","current_price":517.93,"previous_close":520.15,"updated_at":"2025-09-23T16:35:40Z"}
```

## Security & Performance

### Rate Limiting
- **Transactions**: 10 requests per 60 seconds
- **Authentication**: 5 requests per 10 minutes  
- **General API**: 100 requests per hour
- **Search/Browse**: Standard rate limiting

### Caching Strategy
- **Stock Prices**: 30-second TTL with automatic refresh
- **User Sessions**: JWT token validation with 1 hour expiry time

### Architecture Flow
```
Yahoo API → Stocker (Background Worker) → Redis Streams → Processor → PostgreSQL
                                                              ↓
User Holdings ← SSE Stream ← /api/events ← User's Stock Symbols
```

### Conclusion
*If you've read it till this end, consider giving a star!*

*Built with ❤️ using Go, designed for scale and performance by Cheems!*