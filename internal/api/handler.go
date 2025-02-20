package api

import (
	"erply_test/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	router *gin.Engine
	logger logger.LoggerInterface
}

func NewHandler(router *gin.Engine, logger logger.LoggerInterface) *APIHandler {
	return &APIHandler{
		router: router,
		logger: logger,
	}
}

func (h *APIHandler) GetHealth(c *gin.Context) {

	var err error
	if err != nil {
		h.logger.Error("error", err)

	}
	c.JSON(http.StatusOK, gin.H{"status": "good"})
}
