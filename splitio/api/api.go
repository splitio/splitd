package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/api/controllers"
	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/link/client"
)

func Setup(host string, port int, logger logging.LoggerInterface, listenerCfG link.ListenerOptions) (*http.Server, error) {

	router := gin.Default()
	mainAPI := router.Group("/api")

	healthCtrl := controllers.NewHealthController(logger, link.ConsumerOptions{
		Transfer: listenerCfG.Transfer,
		Consumer: client.Options{
			ID:                  strconv.Itoa(os.Getpid()),
			Protocol:            listenerCfG.Protocol,
			ImpressionsFeedback: false,
		},
		Serialization: listenerCfG.Serialization,
	})

	healthCtrl.Register(mainAPI)

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: router,
	}, nil
}
