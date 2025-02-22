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

	sharedCommon "github.com/erply/api-go-wrapper/pkg/api/common"
	"github.com/erply/api-go-wrapper/pkg/api/customers"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type MockCustomerManager struct {
	DeleteCustomerBulkFunc func(
		ctx context.Context,
		ids []map[string]interface{},
		params map[string]string,
	) (customers.DeleteCustomersResponseBulk, error)
}

func (m *MockCustomerManager) DeleteCustomerBulk(
	ctx context.Context,
	ids []map[string]interface{},
	params map[string]string,
) (customers.DeleteCustomersResponseBulk, error) {

	if m.DeleteCustomerBulkFunc != nil {
		return m.DeleteCustomerBulkFunc(ctx, ids, params)
	}

	statusBulk := sharedCommon.StatusBulk{}
	statusBulk.ResponseStatus = "ok"
	bulkResp := customers.DeleteCustomersResponseBulk{
		Status: sharedCommon.Status{ResponseStatus: "ok"},
		BulkItems: []customers.DeleteCustomerResponseBulkItem{
			{
				Status: statusBulk,
			},
			{
				Status: statusBulk,
			},
		},
	}
	return bulkResp, nil
}

func (m *MockCustomerManager) GetCustomers(ctx context.Context, filters map[string]string) (customers.GetCustomersResponse, error) {
	return customers.GetCustomersResponse{}, nil
}
func (m *MockCustomerManager) GetCustomersBulk(ctx context.Context, bulkFilters []map[string]interface{}, baseFilters map[string]string) (customers.GetCustomersResponseBulk, error) {
	return customers.GetCustomersResponseBulk{}, nil
}
func (m *MockCustomerManager) SaveCustomerBulk(ctx context.Context, req []map[string]interface{}, params map[string]string) (customers.SaveCustomerResponseBulk, error) {
	return customers.SaveCustomerResponseBulk{}, nil
}

func TestDeleteCustomers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	testLogger := logger.NewSlogLogger()
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	if err := redisClient.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("Failed to flush redis: %v", err)
	}

	testCache := cache.NewRedisCache(redisClient)

	mockManager := &MockCustomerManager{
		DeleteCustomerBulkFunc: func(
			ctx context.Context,
			ids []map[string]interface{},
			params map[string]string,
		) (customers.DeleteCustomersResponseBulk, error) {
			statusBulk := sharedCommon.StatusBulk{}
			statusBulk.ResponseStatus = "ok"

			bulkResp := customers.DeleteCustomersResponseBulk{
				Status: sharedCommon.Status{ResponseStatus: "ok"},
				BulkItems: []customers.DeleteCustomerResponseBulkItem{
					{
						Status: statusBulk,
					},
					{
						Status: statusBulk,
					},
				},
			}
			return bulkResp, nil
		},
	}
	mockClient := &customers.Client{
		CustomerManager: mockManager,
	}
	handler := api.NewHandler(r, testLogger, mockClient, testCache)

	r.DELETE("/customers/delete", handler.DeleteCustomers)
	reqBody := map[string]interface{}{
		"customerIDs": []int{13380, 13381},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodDelete, "/customers/delete", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Response should be valid JSON")

	assert.Equal(t, "ok", resp["status"], "Expected status 'ok'")
	response, ok := resp["response"].(map[string]interface{})
	assert.True(t, ok, "Expected 'response' field in JSON")

	recordsDeleted, _ := response["recordsDeleted"].(float64)
	assert.Equal(t, 2, int(recordsDeleted), "Expected 2 records to be deleted")
}
