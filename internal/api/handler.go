package api

import (
	"encoding/json"
	"erply_test/internal/logger"
	cache "erply_test/internal/repository"
	"net/http"
	"time"

	"context"

	"github.com/erply/api-go-wrapper/pkg/api"
	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	router      *gin.Engine
	logger      logger.LoggerInterface
	ctx         context.Context
	erplyClient *api.Client
	cache       cache.CacheInterface
}

func NewHandler(router *gin.Engine, logger logger.LoggerInterface, erplyClient *api.Client, cache cache.CacheInterface) *APIHandler {
	return &APIHandler{
		router:      router,
		logger:      logger,
		erplyClient: erplyClient,
		cache:       cache,
	}
}

func (h *APIHandler) GetHealth(c *gin.Context) {

	var err error
	if err != nil {
		h.logger.Error("error", err)

	}
	c.JSON(http.StatusOK, gin.H{"status": "good"})
}

func (h *APIHandler) GetCustomers(c *gin.Context) {
	ctx := c.Request.Context()

	val, err := h.cache.Get(ctx, "customers")
	if err != nil {
		h.logger.Error("error getting from cache", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if val == "" {
		customers, err := h.erplyClient.CustomerManager.GetCustomers(ctx, map[string]string{})
		if err != nil {
			h.logger.Error("error fetching customers", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		customersJSON, err := json.Marshal(customers)
		if err != nil {
			h.logger.Error("error marshalling customers", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := h.cache.Set(ctx, "customers", string(customersJSON), 10*time.Minute); err != nil {
			h.logger.Error("error caching customers", err)
		}

		val = string(customersJSON)
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": val,
	})
}
