package api

import (
	"context"
	"encoding/json"
	"erply_test/internal/logger"
	cache "erply_test/internal/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/erply/api-go-wrapper/pkg/api"
	"github.com/erply/api-go-wrapper/pkg/api/customers"
	"github.com/gin-gonic/gin"
)

// --- Structures for request payload ---

// Request payload for deleting multiple customers by IDs
type DeleteRequest struct {
	CustomerIDs []int `json:"customerIDs"`
}

// Request payload for saving multiple customers
type SaveRequest struct {
	Customers []SaveCustomer `json:"customers"`
}

// A single customer to be saved
type SaveCustomer struct {
	CustomerID  *int   `json:"customerID,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	// Add whatever fields you need that Erply supports
}

type APIHandler struct {
	router      *gin.Engine
	logger      logger.LoggerInterface
	ctx         context.Context
	erplyClient *api.Client
	cache       cache.CacheInterface
}

func NewHandler(
	router *gin.Engine,
	logger logger.LoggerInterface,
	erplyClient *api.Client,
	cache cache.CacheInterface,
) *APIHandler {
	return &APIHandler{
		router:      router,
		logger:      logger,
		erplyClient: erplyClient,
		cache:       cache,
	}
}

// GetHealth godoc
// @Summary     Returns health status
// @Description Simple healthcheck endpoint
// @Tags        health
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Router      /health [get]
func (h *APIHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "good"})
}

// GetCustomers godoc
// @Summary     Fetch Customers
// @Description Get customers from Erply, possibly from cache. Allows query params like "pageNo" and "recordsOnPage".
// @Tags        customers
// @Accept      json
// @Produce     json
// @Param       pageNo        query int false "Page number"
// @Param       recordsOnPage query int false "Records per page"
// @Success     200 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/customers [get]
func (h *APIHandler) GetCustomers(c *gin.Context) {
	ctx, cli := h.init(30)

	pageNoStr := c.Query("pageNo")
	recordsOnPageStr := c.Query("recordsOnPage")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	if recordsOnPageStr == "" {
		recordsOnPageStr = "100"
	}

	cacheKey := "customers_" + pageNoStr + "_" + recordsOnPageStr
	val, err := h.cache.Get(ctx, cacheKey)
	if err != nil {
		h.logger.Error("error getting from cache", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if val == "" {
		// If not found in cache, fetch from Erply
		bulkFilters := []map[string]interface{}{
			{
				"recordsOnPage": recordsOnPageStr,
				"pageNo":        pageNoStr,
			},
		}
		customersResp, err := cli.GetCustomersBulk(ctx, bulkFilters, map[string]string{})
		if err != nil {
			h.logger.Error("error fetching customers", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		customersJSON, err := json.Marshal(customersResp)
		if err != nil {
			h.logger.Error("error marshalling customers", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := h.cache.Set(ctx, cacheKey, string(customersJSON), 10*time.Minute); err != nil {
			h.logger.Error("error caching customers", err)
		}
		val = string(customersJSON)
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": val,
	})
}

// DeleteCustomers godoc
// @Summary     Delete Customers
// @Description Delete one or more customers by their IDs
// @Tags        customers
// @Accept      json
// @Produce     json
// @Param       request body DeleteRequest true "Delete request"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/customers/delete [delete]
func (h *APIHandler) DeleteCustomers(c *gin.Context) {
	ctx, cli := h.init(10)

	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid json for delete request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if len(req.CustomerIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no customer IDs provided"})
		return
	}

	// Build the slice required by DeleteCustomerBulk
	var bulkReq []map[string]interface{}
	for _, id := range req.CustomerIDs {
		bulkReq = append(bulkReq, map[string]interface{}{
			"customerID": strconv.Itoa(id),
		})
	}

	deleteResp, err := cli.DeleteCustomerBulk(ctx, bulkReq, map[string]string{})
	if err != nil {
		h.logger.Error("error deleting customers", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "response": deleteResp})
}

// SaveCustomers godoc
// @Summary     Save Customers
// @Description Create or update customers in Erply
// @Tags        customers
// @Accept      json
// @Produce     json
// @Param       request body SaveRequest true "Customers to save"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/customers/save [post]
func (h *APIHandler) SaveCustomers(c *gin.Context) {
	ctx, cli := h.init(30)

	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid json for save request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if len(req.Customers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no customers to save"})
		return
	}

	var bulk []map[string]interface{}
	for _, cust := range req.Customers {
		m := map[string]interface{}{}
		if cust.CustomerID != nil {
			m["customerID"] = *cust.CustomerID
		}
		if cust.FirstName != "" {
			m["firstName"] = cust.FirstName
		}
		if cust.LastName != "" {
			m["lastName"] = cust.LastName
		}
		if cust.CompanyName != "" {
			m["companyName"] = cust.CompanyName
		}
		if cust.Email != "" {
			m["email"] = cust.Email
		}
		if cust.Phone != "" {
			m["phone"] = cust.Phone
		}
		bulk = append(bulk, m)
	}

	resp, err := cli.SaveCustomerBulk(ctx, bulk, map[string]string{})
	if err != nil {
		h.logger.Error("error saving customers", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) init(ttl int16) (context.Context, customers.Manager) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(ttl))
	cli := h.erplyClient.CustomerManager
	return ctx, cli
}
