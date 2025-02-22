package test

import (
	"bytes"
	"context"
	"encoding/json"
	"erply_test/internal/api"
	"erply_test/internal/logger"
	cache "erply_test/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erply/api-go-wrapper/pkg/api"
	sharedCommon "github.com/erply/api-go-wrapper/pkg/api/common"
	"github.com/erply/api-go-wrapper/pkg/api/customers"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------
// 1) Mock Customer Manager
// ---------------------------------------

type MockCustomerManager struct {
	DeleteCustomerBulkFunc func(
		ctx context.Context,
		ids []map[string]interface{},
		params map[string]string,
	) (sharedCommon.DeleteResponseBulk, error)
}

func (m *MockCustomerManager) DeleteCustomerBulk(
	ctx context.Context,
	ids []map[string]interface{},
	params map[string]string,
) (sharedCommon.DeleteResponseBulk, error) {
	if m.DeleteCustomerBulkFunc != nil {
		return m.DeleteCustomerBulkFunc(ctx, ids, params)
	}
	// Default response if no custom func is provided
	return sharedCommon.DeleteResponseBulk{
		Status: sharedCommon.Status{
			ResponseStatus:   "ok",
			RecordsInRequest: len(ids),
			RecordsDeleted:   len(ids),
		},
		Results: []sharedCommon.DeleteResponse{
			{DeletedID: 13380, ErrorCode: 0},
			{DeletedID: 13381, ErrorCode: 0},
		},
	}, nil
}

// Required to satisfy the interface, no-op implementations
func (m *MockCustomerManager) GetCustomers(ctx context.Context, filters map[string]string) (customers.GetCustomersResponse, error) {
	return customers.GetCustomersResponse{}, nil
}
func (m *MockCustomerManager) GetCustomersBulk(ctx context.Context, bulkFilters []map[string]interface{}, baseFilters map[string]string) (customers.GetCustomersResponseBulk, error) {
	return customers.GetCustomersResponseBulk{}, nil
}
func (m *MockCustomerManager) SaveCustomerBulk(ctx context.Context, req []map[string]interface{}, params map[string]string) (sharedCommon.MultiRequestResponse, error) {
	return sharedCommon.MultiRequestResponse{}, nil
}

// ---------------------------------------
// 2) Test for DeleteCustomers
// ---------------------------------------

func TestDeleteCustomers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup a test recorder and context
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	// Initialize logger and Redis cache
	testLogger := logger.NewSlogLogger()
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Update as needed
	})
	defer redisClient.Close()

	// Flush Redis for clean state
	if err := redisClient.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("Failed to flush redis: %v", err)
	}

	testCache := cache.NewRedisCache(redisClient)

	// Mock DeleteCustomerBulk response
	mockManager := &MockCustomerManager{
		DeleteCustomerBulkFunc: func(
			ctx context.Context,
			ids []map[string]interface{},
			params map[string]string,
		) (sharedCommon.DeleteResponseBulk, error) {
			return sharedCommon.DeleteResponseBulk{
				Status: sharedCommon.Status{
					ResponseStatus:   "ok",
					RecordsInRequest: len(ids),
					RecordsDeleted:   len(ids),
				},
				Results: []sharedCommon.DeleteResponse{
					{DeletedID: 13380, ErrorCode: 0},
					{DeletedID: 13381, ErrorCode: 0},
				},
			}, nil
		},
	}

	// Mock API client with the fake manager
	mockErplyClient := &api.Client{
		CustomerManager: mockManager,
	}

	// Create handler with mock client
	handler := api.NewHandler(r, testLogger, mockErplyClient, testCache)

	// Register the route for testing
	r.DELETE("/customers/delete", handler.DeleteCustomers)

	// Prepare the request body
	reqBody := map[string]interface{}{
		"customerIDs": []int{13380, 13381},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodDelete, "/customers/delete", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	c.Request = req
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Response should be valid JSON")

	// Check for expected response keys
	assert.Equal(t, "ok", resp["status"], "Expected status 'ok'")

	// Check records deleted count
	response, ok := resp["response"].(map[string]interface{})
	assert.True(t, ok, "Expected 'response' field in JSON")

	recordsDeleted, _ := response["recordsDeleted"].(float64)
	assert.Equal(t, 2, int(recordsDeleted), "Expected 2 records to be deleted")
}
