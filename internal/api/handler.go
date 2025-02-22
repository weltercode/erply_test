package api

import (
	"context"
	"encoding/json"
	"erply_test/internal/logger"
	cache "erply_test/internal/repository"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/erply/api-go-wrapper/pkg/api/customers"
	"github.com/gin-gonic/gin"
)

type DeleteRequest struct {
	CustomerIDs []interface{} `json:"customerIDs"`
}

type SaveRequest struct {
	Customers []SaveCustomer `json:"customers"`
}

type SaveCustomer struct {
	CustomerID  *int   `json:"customerID,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

type APIHandler struct {
	router          *gin.Engine
	logger          logger.LoggerInterface
	customerManager customers.Manager
	cache           cache.CacheInterface
}

func NewHandler(
	router *gin.Engine,
	logger logger.LoggerInterface,
	customerManager customers.Manager,
	cache cache.CacheInterface,
) *APIHandler {
	return &APIHandler{
		router:          router,
		logger:          logger,
		customerManager: customerManager,
		cache:           cache,
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
// @Security    ApiKeyAuth
func (h *APIHandler) GetCustomers(c *gin.Context) {
	ctx := h.init(30)

	pageNoStr := c.Query("pageNo")
	recordsOnPageStr := c.Query("recordsOnPage")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	if recordsOnPageStr == "" {
		recordsOnPageStr = "100"
	}

	//cacheKey := "customers_" + pageNoStr + "_" + recordsOnPageStr
	cacheKey := "customers"
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
		customersResp, err := h.customerManager.GetCustomersBulk(ctx, bulkFilters, map[string]string{})
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
// @Security    ApiKeyAuth
func (h *APIHandler) DeleteCustomers(c *gin.Context) {
	ctx := h.init(10)

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

	var bulkReq []map[string]interface{}
	for _, id := range req.CustomerIDs {
		var idStr string
		switch v := id.(type) {
		case float64:
			idStr = strconv.Itoa(int(v))
		case string:
			idStr = v
		default:
			h.logger.Error("invalid customer ID type", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID type"})
			return
		}

		h.logger.Info("Deleting customerID: " + idStr)
		bulkReq = append(bulkReq, map[string]interface{}{
			"customerID": idStr,
		})
	}

	deleteResp, err := h.customerManager.DeleteCustomerBulk(ctx, bulkReq, map[string]string{})
	if err != nil {
		//`Invalid classifier ID, there is no such item. (Attribute "errorField" indicates the invalid input parameter.)`, from Erply-go-wrapper
		if strings.Contains(err.Error(), "1011") {
			h.logger.Warn("Customer already deleted or invalid ID")
			c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Some customers were already deleted.", "response": deleteResp})
			return
		}
		h.logger.Error("error deleting customers", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "customerIDs": req.CustomerIDs})
		return
	}
	cacheKey := "customers"
	h.cache.Delete(ctx, cacheKey)

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
// @Security    ApiKeyAuth
func (h *APIHandler) SaveCustomers(c *gin.Context) {
	ctx := h.init(30)

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

	resp, err := h.customerManager.SaveCustomerBulk(ctx, bulk, map[string]string{})
	if err != nil {
		h.logger.Error("error saving customers", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cacheKey := "customers"
	h.cache.Delete(ctx, cacheKey)

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) init(ttl int16) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(ttl))
	return ctx
}
