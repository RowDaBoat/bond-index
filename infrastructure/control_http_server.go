package infrastructure

import (
	"fmt"
	"net/http"
	"os"

	"bond/actions"

	"github.com/gin-gonic/gin"
)

type ControlHttpServer struct {
	syncRequests chan actions.SyncRequest
	router       *gin.Engine
}

func NewControlHttpServer(syncRequests chan actions.SyncRequest) *ControlHttpServer {
	router := gin.Default()

	server := &ControlHttpServer{
		syncRequests: syncRequests,
		router:       router,
	}

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/sync", server.handleSync)

	return server
}

func (s *ControlHttpServer) Start(address string) {
	fmt.Printf("Starting control HTTP server on %s\n", address)
	if err := s.router.Run(address); err != nil {
		fmt.Printf("Control HTTP server error: %v\n", err)
		os.Exit(1)
	}
}

func (s *ControlHttpServer) handleSync(c *gin.Context) {
	reply := make(actions.SyncRequest, 1)
	s.syncRequests <- reply
	blockHeight := <-reply

	c.JSON(http.StatusOK, gin.H{
		"synced_to_block": blockHeight,
	})
}
