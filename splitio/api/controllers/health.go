package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/link/client"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
)

type HealthCheckController struct {
	connParams link.ConsumerOptions
	logger     logging.LoggerInterface
}

func (c *HealthCheckController) Register(router gin.IRouter) {
	router.GET("/health", c.isHealthy)
	router.GET("/ready", c.isReady)
	router.GET("/checkFlag", c.checkFlag)
}

func (c *HealthCheckController) isHealthy(ctx *gin.Context) {
	ctx.Status(200)
}

func (c *HealthCheckController) isReady(ctx *gin.Context) {
	conn, err := transfer.NewClientConn(c.logger, &c.connParams.Transfer)
	if conn != nil {
		defer conn.Shutdown()
	}
	if err != nil {
		ctx.AbortWithError(500, fmt.Errorf("error creating raw connection: %w", err))
		return
	}

	serial, err := serializer.Setup(c.connParams.Serialization)
	if err != nil {
		ctx.AbortWithError(500, fmt.Errorf("error setting up serializer: %w", err))
		return
	}

	_, err = client.New(c.logger, conn, serial, c.connParams.Consumer)
	if err != nil {
		ctx.AbortWithError(500, fmt.Errorf("error setting up client: %w", err))
		return
	}

	ctx.Status(200)
}

func (c *HealthCheckController) checkFlag(ctx *gin.Context) {

	splitName := ctx.Request.URL.Query().Get("flag")
	conn, err := transfer.NewClientConn(c.logger, &c.connParams.Transfer)
	if conn != nil {
		defer conn.Shutdown()
	}
	if err != nil {
		ctx.AbortWithError(500, fmt.Errorf("error creating raw connection: %w", err))
		return
	}

	serial, err := serializer.Setup(c.connParams.Serialization)
	if err != nil {
		ctx.AbortWithError(500, fmt.Errorf("error setting up serializer: %w", err))
		return
	}

	rpcClient, err := client.New(c.logger, conn, serial, c.connParams.Consumer)
	if err != nil {
		ctx.AbortWithError(500, fmt.Errorf("error setting up client: %w", err))
		return
	}

	result, err := rpcClient.Split(splitName)
	if err != nil {
		ctx.AbortWithError(500, fmt.Errorf("error issuing RPC: %w", err))
		return
	}

	if result == nil {
		ctx.AbortWithStatus(404)
		return
	}

	ctx.JSON(200, SplitViewDTO(*result))
}

func NewHealthController(logger logging.LoggerInterface, connParams link.ConsumerOptions) *HealthCheckController {
	return &HealthCheckController{
		connParams: connParams,
		logger:     logger,
	}
}
