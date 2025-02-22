package main

import (
	"erply_test/internal/app"
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// @title Erply customers API test wrapper
// @version 1.0
// @description This is an API for managing Erply customers. It allows you to fetch customers, save them, and delete them.\n It use https://github.com/erply/api-go-wrapper/ to interact with Erply API. \n 127.0.0.1:3000 if you run it locally, or :8080 if you run it in Docker
// @host 127.0.0.1:3000
// @BasePath /
// @healthcheck /health
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
// @security ApiKeyAuth

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found z, using system environment variables")
	}

	cfg := app.Config{}

	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	application := app.CreateApp(&cfg)
	application.Run()
}
