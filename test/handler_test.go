package test

import (
	"bytes"
	"context"
	"erply_test/internal/api"
	"erply_test/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sharedCommon "github.com/erply/api-go-wrapper/pkg/api/common"
	"github.com/erply/api-go-wrapper/pkg/api/customers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCustomerManager struct {
	mock.Mock
}

func (m *MockCustomerManager) GetCustomersBulk(ctx context.Context, filters []map[string]interface{}, opts map[string]string) (customers.GetCustomersResponseBulk, error) {
	args := m.Called(ctx, filters, opts)
	if result, ok := args.Get(0).(customers.GetCustomersResponseBulk); ok {
		return result, args.Error(1)
	}
	return customers.GetCustomersResponseBulk{}, args.Error(1)
}

func (m *MockCustomerManager) DeleteCustomerBulk(ctx context.Context, bulk []map[string]interface{}, opts map[string]string) (customers.DeleteCustomersResponseBulk, error) {
	args := m.Called(ctx, bulk, opts)
	if result, ok := args.Get(0).(customers.DeleteCustomersResponseBulk); ok {
		return result, args.Error(1)
	}
	return customers.DeleteCustomersResponseBulk{}, args.Error(1)
}

func (m *MockCustomerManager) SaveCustomerBulk(ctx context.Context, bulk []map[string]interface{}, opts map[string]string) (customers.SaveCustomerResponseBulk, error) {
	args := m.Called(ctx, bulk, opts)
	if result, ok := args.Get(0).(customers.SaveCustomerResponseBulk); ok {
		return result, args.Error(1)
	}
	return customers.SaveCustomerResponseBulk{}, args.Error(1)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSaveCustomers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockManager := new(MockCustomerManager)
	mockCache := new(MockCache)
	handler := api.NewHandler(gin.Default(), logger.NewSlogLogger(), mockManager, mockCache)

	r := gin.Default()
	r.POST("/api/customers/save", handler.SaveCustomers)

	mockManager.On("SaveCustomerBulk", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mockCache.On("Delete", mock.Anything, mock.AnythingOfType("[]string")).Return(nil)

	body := []byte(`{"customers": [{"firstName": "Anna", "lastName": "Taylor", "companyName": "Company 1", "email": "}]}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/customers/save", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetCustomers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockManager := new(MockCustomerManager)
	mockCache := new(MockCache)
	handler := api.NewHandler(gin.Default(), logger.NewSlogLogger(), mockManager, mockCache)

	r := gin.Default()
	r.GET("/api/customers", handler.GetCustomers)
	mockCache.On("Get", mock.Anything, "customers").Return("", nil)
	mockManager.On("GetCustomersBulk", mock.Anything, mock.Anything, mock.Anything).
		Return(customers.GetCustomersResponseBulk{
			Status: sharedCommon.Status{ResponseStatus: "ok"},
			BulkItems: []customers.GetCustomersResponseBulkItem{
				{
					Status: sharedCommon.StatusBulk{
						RequestName: "getCustomers",
						RequestID:   "12345",
						Status: sharedCommon.Status{
							ResponseStatus: "ok",
						},
					},
					Customers: []customers.Customer{
						{
							ID:          123,
							CompanyName: "Customer 123",
						},
						{
							ID:          124,
							CompanyName: "Customer 124",
						},
					},
				},
				{
					Status: sharedCommon.StatusBulk{
						RequestName: "getCustomers",
						RequestID:   "12346",
						Status: sharedCommon.Status{
							ResponseStatus: "ok",
						},
					},
					Customers: []customers.Customer{
						{
							ID:          125,
							CompanyName: "Customer 125",
						},
					},
				},
			},
		}, nil)

	mockCache.On("Set", mock.Anything, "customers", mock.Anything, mock.Anything).Return(nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/customers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestDeleteCustomers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockManager := new(MockCustomerManager)
	mockCache := new(MockCache)
	handler := api.NewHandler(gin.Default(), logger.NewSlogLogger(), mockManager, mockCache)

	r := gin.Default()
	r.DELETE("/api/customers/delete", handler.DeleteCustomers)

	mockManager.On("DeleteCustomerBulk", mock.Anything, mock.Anything, mock.Anything).Return(customers.DeleteCustomersResponseBulk{}, nil)
	mockCache.On("Delete", mock.Anything, mock.AnythingOfType("[]string")).Return(nil)

	body := []byte(`{"customerIDs": [1, 2, 3]}`)
	req, _ := http.NewRequest(http.MethodDelete, "/api/customers/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
