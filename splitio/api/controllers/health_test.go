package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resp := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(resp)
	group := router.Group("/api")
	controller := NewHealthController(logging.NewLogger(nil), link.ConsumerOptions{})

	controller.Register(group)

	ctx.Request, _ = http.NewRequest(http.MethodGet, "/api/health", nil)
	router.ServeHTTP(resp, ctx.Request)

	assert.Equal(t, 200, resp.Code)
}
