package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"erply_test/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No API key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"API key is required"}`,
		},
		{
			name:           "Invalid API key",
			apiKey:         "invalid_key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid API key"}`,
		},
		{
			name:           "Valid API key",
			apiKey:         "valid_key",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.APIKeyAuthMiddleware("valid_key"))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-KEY", tt.apiKey)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
