package app

import (
	"context"
	mongodb "erply_test/internal/database"
	"erply_test/internal/logger"
	"fmt"
	"net/http"

	"github.com/erply/api-go-wrapper/pkg/api"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	config      *Config
	router      *gin.Engine
	logger      logger.LoggerInterface
	db          *mongo.Database
	ctx         context.Context
	erplyClient *api.Client
}

type Config struct {
	MONGO_URI         string `env:"MONGO_URI"`
	DbName            string `env:"MONGO_DB"`
	DbUser            string `env:"MONGO_USER"`
	DbPass            string `env:"MONGO_PASSWORD"`
	AppPort           string `env:"APP_PORT"`
	AppHost           string `env:"APP_HOST"`
	ERPLY_USER_NAME   string `env:"ERPLY_USER_NAME"`
	ERPLY_USER_PASS   string `env:"ERPLY_USER_PASS"`
	ERPLY_CLIENT_CODE string `env:"ERPLY_CLIENT_CODE"`
}

func CreateApp(config *Config) *App {
	fmt.Println()
	logger := logger.NewSlogLogger()
	db := mongodb.ConnectDB(mongodb.ConnectionConfig{
		URI:    config.MONGO_URI,
		DbName: config.DbName,
		User:   config.DbUser,
		Pass:   config.DbPass,
	}, logger)

	ctx := context.Background()

	var z = map[string]string{}

	cli, err := api.NewClientFromCredentials(config.ERPLY_USER_NAME, config.ERPLY_USER_PASS, config.ERPLY_CLIENT_CODE, nil)
	cli.CustomerManager.SaveCustomer(ctx, z)
	if err != nil {
		panic(err)
	}

	if err := db.Client().Ping(ctx, nil); err != nil {
		logger.Error("Fail to connect DB", err)
	} else {
		logger.Info("Database connected", err)
	}

	return &App{
		config:      config,
		router:      gin.Default(),
		db:          db,
		logger:      logger,
		erplyClient: cli,
	}
}

func (app *App) Run() {
	defer app.Shutdown()
	app.router.GET("/", getHome)
	app.router.GET("/health", getHealth)
	app.logger.Info("App Running")
	app.logger.Info(app.config.AppHost + ":" + app.config.AppPort)
	app.router.Run(app.config.AppHost + ":" + app.config.AppPort)
}

func (app *App) Shutdown() {
	if app.db != nil {
		app.logger.Info("Closing database connection...")
		app.db.Client().Disconnect(app.ctx)
	}
}

func getHome(c *gin.Context) {
	c.IndentedJSON(http.StatusCreated, "hello")
}

func getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
