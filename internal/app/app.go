package app

import (
	"context"
	hapi "erply_test/internal/api"
	"erply_test/internal/logger"
	"erply_test/internal/middleware"
	"fmt"
	"net/http"

	"github.com/erply/api-go-wrapper/pkg/api"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type App struct {
	config      *Config
	router      *gin.Engine
	logger      logger.LoggerInterface
	cache       *redis.Client
	ctx         context.Context
	erplyClient *api.Client
	handler     *hapi.APIHandler
}

type Config struct {
	RedisAddr         string `env:"REDIS_ADDR"`
	AppPort           string `env:"APP_PORT"`
	AppHost           string `env:"APP_HOST"`
	ERPLY_USER_NAME   string `env:"ERPLY_USER_NAME"`
	ERPLY_USER_PASS   string `env:"ERPLY_USER_PASS"`
	ERPLY_CLIENT_CODE string `env:"ERPLY_CLIENT_CODE"`
	ApiKey            string `env:"API_KEY"`
}

func CreateApp(config *Config) *App {
	fmt.Println()
	logger := logger.NewSlogLogger()
	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	} else {
		logger.Info("Connected to Redis!", nil)
	}

	var z = map[string]string{}

	cli, err := api.NewClientFromCredentials(config.ERPLY_USER_NAME, config.ERPLY_USER_PASS, config.ERPLY_CLIENT_CODE, nil)
	cli.CustomerManager.SaveCustomer(ctx, z)
	if err != nil {
		panic(err)
	}

	var router = gin.Default()

	return &App{
		config:      config,
		router:      router,
		cache:       redisClient,
		logger:      logger,
		erplyClient: cli,
		handler:     hapi.NewHandler(router, logger),
	}
}

func (app *App) Run() {
	defer app.Shutdown()

	// ==========  Public routes  ==========
	app.router.GET("/health", app.handler.GetHealth)

	// ==========  Protected routes  ==========
	protected := app.router.Group("/api")
	protected.Use(middleware.APIKeyAuthMiddleware(app.config.ApiKey))
	{
		protected.GET("/customers", func(c *gin.Context) {
			val, err := app.cache.Get(app.ctx, "customers").Result()
			if err == redis.Nil {
				val = "not found in redis"
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"msg":         "You have access to protected customers endpoint",
				"redis_value": val,
			})
		})
	}
	app.logger.Info("App Running")
	app.logger.Info(app.config.AppHost + ":" + app.config.AppPort)
	app.router.Run(app.config.AppHost + ":" + app.config.AppPort)
}

func (app *App) Shutdown() {
	if err := app.cache.Close(); err != nil {
		app.logger.Error("Error closing Redis client", err)
	}
}
