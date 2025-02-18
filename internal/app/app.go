package app

import (
	"context"
	mongodb "erply_test/internal/database"
	"erply_test/internal/logger"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	config *Config
	router *gin.Engine
	logger logger.LoggerInterface
	db     *mongo.Database
	ctx    context.Context
}

type Config struct {
	DbIp    string `env:"MONOGO_DB_IP"`
	DbPort  string `env:"MONGIO_DB_PORT"`
	DbName  string `env:"MONGO_DB_NAME"`
	DbUser  string `env:"MONGO_USER"`
	DbPass  string `env:"MONGO_PASSWORD"`
	AppPort string `env:"APP_PORT"`
	AppHost string `env:"APP_HOST"`
}

func CreateApp(config *Config) *App {
	logger := logger.NewSlogLogger()
	db := mongodb.ConnectDB(mongodb.ConnectionConfig{
		Host:   config.DbIp,
		Port:   config.DbPort,
		DbName: config.DbName,
		User:   config.DbUser,
		Pass:   config.DbPass,
	}, logger)

	ctx := context.Background()
	if err := db.Client().Ping(ctx, nil); err != nil {
		logger.Error("Fail to connect DB", err)
	} else {
		logger.Info("Database connected", err)
	}

	return &App{
		config: config,
		router: gin.Default(),
		db:     db,
		logger: logger,
	}
}

func (app *App) Run() {
	defer app.Shutdown()
	app.logger.Info("App Running")
	app.router.Run(":" + app.config.AppPort)
}

func (app *App) Shutdown() {
	if app.db != nil {
		app.logger.Info("Closing database connection...")
		app.db.Client().Disconnect(app.ctx)
	}
}
