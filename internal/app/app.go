package app

import (
	"context"
	hapi "erply_test/internal/api"
	"erply_test/internal/logger"
	"erply_test/internal/middleware"
	cache "erply_test/internal/repository"
	"fmt"

	_ "erply_test/docs"

	"github.com/gin-contrib/cors"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/erply/api-go-wrapper/pkg/api"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type App struct {
	config      *Config
	router      *gin.Engine
	logger      logger.LoggerInterface
	cache       cache.CacheInterface
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

	cache := cache.NewRedisCache(redisClient)

	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	} else {
		logger.Info("Connected to Redis!", nil)
	}

	erplyClient, err := api.NewClientFromCredentials(config.ERPLY_USER_NAME, config.ERPLY_USER_PASS, config.ERPLY_CLIENT_CODE, nil)

	if err != nil {
		panic(err)
	}

	var router = gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-KEY"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	return &App{
		config:      config,
		router:      router,
		cache:       cache,
		logger:      logger,
		erplyClient: erplyClient,
		handler:     hapi.NewHandler(router, logger, erplyClient.CustomerManager, cache),
	}
}

func (app *App) Run() {
	defer app.Shutdown()

	// ==========  Public routes  ==========
	app.router.GET("/health", app.handler.GetHealth)
	app.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ==========  Protected routes  ==========
	protected := app.router.Group("/api")
	protected.Use(middleware.APIKeyAuthMiddleware(app.config.ApiKey))
	{
		protected.GET("/customers", app.handler.GetCustomers)
		protected.DELETE("/customers/delete", app.handler.DeleteCustomers)
		protected.POST("/customers/save", app.handler.SaveCustomers)
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
