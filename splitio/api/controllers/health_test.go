package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/splitio/splitd/splitio/sdk/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {

	gin.SetMode(gin.TestMode)

	resp := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(resp)
	group := router.Group("/api")

	logger := logging.NewLogger(nil)

	controller := NewHealthController(logger, link.DefaultConsumerOptions())
	controller.Register(group)

	ctx.Request, _ = http.NewRequest(http.MethodGet, "/api/health", nil)
	router.ServeHTTP(resp, ctx.Request)
	assert.Equal(t, 200, resp.Code)
}

func TestReadinessOk(t *testing.T) {

	gin.SetMode(gin.TestMode)

	resp := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(resp)
	group := router.Group("/api")

	logger := logging.NewLogger(nil)

	var sdkMock mocks.SDKMock

	listenerCfg := link.DefaultListenerOptions()
	listenerCfg.Transfer.ConnType = transfer.ConnTypeUnixStream
	listenerCfg.Transfer.Address = fmt.Sprintf("%s/health_test_%d", os.TempDir(), os.Getpid())
	_, shutdown, err := link.Listen(logger, &sdkMock, &listenerCfg)
	assert.Nil(t, err)
	defer shutdown()

	consumerCfg := link.DefaultConsumerOptions()
	consumerCfg.Transfer = listenerCfg.Transfer

	controller := NewHealthController(logging.NewLogger(nil), consumerCfg)
	controller.Register(group)

	ctx.Request, _ = http.NewRequest(http.MethodGet, "/api/ready", nil)
	router.ServeHTTP(resp, ctx.Request)
	assert.Equal(t, 200, resp.Code)

}

func TestReadinessFailToConnect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resp := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(resp)
	group := router.Group("/api")

	logger := logging.NewLogger(nil)

	var sdkMock mocks.SDKMock

	listenerCfg := link.DefaultListenerOptions()
	listenerCfg.Transfer.ConnType = transfer.ConnTypeUnixStream
	listenerCfg.Transfer.Address = fmt.Sprintf("%s/health_test_%d", os.TempDir(), os.Getpid())
	_, shutdown, err := link.Listen(logger, &sdkMock, &listenerCfg)
	assert.Nil(t, err)
	defer shutdown()

	// by leaving the consumer options with default values, the socket descriptor will not be found
	// on the FS. Connection will fail causing 500 to be returned
	consumerCfg := link.DefaultConsumerOptions()

	controller := NewHealthController(logging.NewLogger(nil), consumerCfg)
	controller.Register(group)

	ctx.Request, _ = http.NewRequest(http.MethodGet, "/api/ready", nil)
	router.ServeHTTP(resp, ctx.Request)
	assert.Equal(t, 500, resp.Code)
}

func TestCheckFlagOk(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resp := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(resp)
	group := router.Group("/api")

	logger := logging.NewLogger(nil)

	var sdkMock mocks.SDKMock
	sdkMock.On("Split", "some-flag").
		Return(&sdk.SplitView{
			Name:             "some-flag",
			TrafficType:      "tt",
			Killed:           true,
			Treatments:       []string{"a", "b"},
			ChangeNumber:     123,
			Configs:          map[string]string{"qwe": "rty"},
			DefaultTreatment: "b",
			Sets:             []string{"s1", "s2"},
		}, nil).
		Once()

	listenerCfg := link.DefaultListenerOptions()
	listenerCfg.Transfer.ConnType = transfer.ConnTypeUnixStream
	listenerCfg.Transfer.Address = fmt.Sprintf("%s/health_test_%d", os.TempDir(), os.Getpid())
	_, shutdown, err := link.Listen(logger, &sdkMock, &listenerCfg)
	assert.Nil(t, err)
	defer shutdown()

	consumerCfg := link.DefaultConsumerOptions()
	consumerCfg.Transfer = listenerCfg.Transfer

	controller := NewHealthController(logging.NewLogger(nil), consumerCfg)
	controller.Register(group)

	ctx.Request, _ = http.NewRequest(http.MethodGet, "/api/checkFlag", nil)
	qp := ctx.Request.URL.Query()
	qp.Add("flag", "some-flag")
	ctx.Request.URL.RawQuery = qp.Encode()
	router.ServeHTTP(resp, ctx.Request)
	assert.Equal(t, 200, resp.Code)

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	var sv SplitViewDTO
	assert.Nil(t, json.Unmarshal(body, &sv))

	assert.Equal(t, SplitViewDTO{
		Name:             "some-flag",
		TrafficType:      "tt",
		Killed:           true,
		Treatments:       []string{"a", "b"},
		ChangeNumber:     123,
		Configs:          map[string]string{"qwe": "rty"},
		DefaultTreatment: "b",
		Sets:             []string{"s1", "s2"},
	}, sv)
}

func TestCheckFlagNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resp := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(resp)
	group := router.Group("/api")

	logger := logging.NewLogger(nil)

	var sdkMock mocks.SDKMock
	sdkMock.On("Split", "some-flag").
		Return((*sdk.SplitView)(nil), sdk.ErrSplitNotFound).
		Once()

	listenerCfg := link.DefaultListenerOptions()
	listenerCfg.Transfer.ConnType = transfer.ConnTypeUnixStream
	listenerCfg.Transfer.Address = fmt.Sprintf("%s/health_test_%d", os.TempDir(), os.Getpid())
	_, shutdown, err := link.Listen(logger, &sdkMock, &listenerCfg)
	assert.Nil(t, err)
	defer shutdown()

	consumerCfg := link.DefaultConsumerOptions()
	consumerCfg.Transfer = listenerCfg.Transfer

	controller := NewHealthController(logging.NewLogger(nil), consumerCfg)
	controller.Register(group)

	ctx.Request, _ = http.NewRequest(http.MethodGet, "/api/checkFlag", nil)
	qp := ctx.Request.URL.Query()
	qp.Add("flag", "some-flag")
	ctx.Request.URL.RawQuery = qp.Encode()
	router.ServeHTTP(resp, ctx.Request)
	assert.Equal(t, 404, resp.Code)
}

func TestCheckFlagConnError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resp := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(resp)
	group := router.Group("/api")

	logger := logging.NewLogger(nil)

	var sdkMock mocks.SDKMock
	sdkMock.On("Split", "some-flag").
		Return((*sdk.SplitView)(nil), nil).
		Once()

	listenerCfg := link.DefaultListenerOptions()
	listenerCfg.Transfer.ConnType = transfer.ConnTypeUnixStream
	listenerCfg.Transfer.Address = fmt.Sprintf("%s/health_test_%d", os.TempDir(), os.Getpid())
	_, shutdown, err := link.Listen(logger, &sdkMock, &listenerCfg)
	assert.Nil(t, err)
	defer shutdown()

	consumerCfg := link.DefaultConsumerOptions()

	controller := NewHealthController(logging.NewLogger(nil), consumerCfg)
	controller.Register(group)

	ctx.Request, _ = http.NewRequest(http.MethodGet, "/api/checkFlag", nil)
	qp := ctx.Request.URL.Query()
	qp.Add("flag", "some-flag")
	ctx.Request.URL.RawQuery = qp.Encode()
	router.ServeHTTP(resp, ctx.Request)
	assert.Equal(t, 500, resp.Code)
}
