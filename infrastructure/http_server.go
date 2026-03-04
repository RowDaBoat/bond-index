package infrastructure

import (
	"fmt"
	"net/http"
	"os"

	"bond/actions"

	"github.com/gin-gonic/gin"
)

type QueryHttpServer struct {
	nameResolver *actions.NameResolver
	router       *gin.Engine
}

func NewQueryHttpServer(nameResolver *actions.NameResolver) *QueryHttpServer {
	router := gin.Default()

	server := &QueryHttpServer{
		nameResolver: nameResolver,
		router:       router,
	}

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/name/:name", server.handleNameQuery)

	return server
}

func (s *QueryHttpServer) Start(address string) {
	fmt.Printf("Starting query HTTP server on %s\n", address)
	if err := s.router.Run(address); err != nil {
		fmt.Printf("Query HTTP server error: %v\n", err)
		os.Exit(1)
	}
}

func (s *QueryHttpServer) handleNameQuery(c *gin.Context) {
	name := c.Param("name")
	result, err := s.nameResolver.Resolve(name)

	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
