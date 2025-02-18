package main

import (
	"erply_test/internal/app"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}
	app := app.CreateApp(&app.Config{
		DbIp:    os.Getenv("DATABASE_IP"),
		DbPort:  os.Getenv("DATABASE_PORT"),
		DbName:  os.Getenv("DATABASE_NAME"),
		DbUser:  os.Getenv("DATABASE_USER"),
		DbPass:  os.Getenv("DATABASE_PASS"),
		AppPort: os.Getenv("APP_PORT"),
	})
	app.Run()
}
