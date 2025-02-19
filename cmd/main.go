package main

import (
	"erply_test/internal/app"
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

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
