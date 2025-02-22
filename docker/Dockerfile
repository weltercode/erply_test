FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
COPY .env .env

RUN go build -o /app/main ./cmd/main.go
RUN go install github.com/swaggo/swag/cmd/swag@latest

RUN swag init -g ./cmd/main.go  

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main .
RUN chmod +x /app/main

EXPOSE 8080
CMD ["./main"]