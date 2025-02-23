# Erply Test Project

This project is a Go application designed to interact with the Erply API.

## Prerequisites
- Go 1.16 or later
- An Erply API account

## Installation
Clone the repository:
```sh
    git clone https://github.com/yourusername/erply_test.git
    cd erply_test
    go mod tidy
```
## Configuration
Create an `.env` file in the root directory and add your Erply API credentials:
```
ERPLY_CLIENT_CODE=your_client_code
ERPLY_USER_NAME=your_erply_username
ERPLY_USER_PASS=your_erply_password
```

Add ```API_KEY``` for secure this API access

The are 3 version of .env files in project:
1) erply_test/.env - used for local development
2) erply_test/docker/.env is used in docker
3) erply_test/.env.example is just for example

## Usage
### Production
```sh
cd .\docker\
docker compose build --no-cache
docker compose up
```
Server will be at APP_HOST:APP_PORT from .env
Default: ```127.0.0.1:8080```

### For development
Run docker to use redis
```sh
cd .\docker\
docker compose build --no-cache
docker compose up
```
Run main.go
```sh
go run cmd/main.go
```

From command line
try to check that api is ready
```sh
curl -i http://127.0.0.1:3000/health
```

try to use API 
```sh
curl -X DELETE \
  -H "Content-Type: application/json" \
  -H "x-api-key: YOUR_API_KEY_FROM_ENV" \
  -d '{
    "customerIDs": [13380, 13381, 13382]
  }' \
  "http://localhost:3000/api/customers/delete"
```
```sh
curl -X GET -H "X-API-KEY: YOUR_API_KEY_FROM_ENV" http://127.0.0.1:3000/api/customers?pageNo=1&recordsOnPage=50
```

From project root (NB! Test json data file located in /json dir ```@json/customers_save.json```)
```sh
curl -X POST -H "Content-Type: application/json" -H "x-api-key: YOUR_API_KEY_FROM_ENV" -d @json/customers_save.json "http://127.0.0.1:3000/api/customers/save"
```

## Test
```sh
go test -v ./test
```

## Swagger Docs
```sh
swag init -g cmd/main.go  
```