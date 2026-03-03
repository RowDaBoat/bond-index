package infrastructure

import (
	"fmt"
	"net/http"

	"bond/actions"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	nameResolver *actions.NameResolver
	router       *gin.Engine
}

func NewHttpServer(nameResolver *actions.NameResolver) *HttpServer {
	router := gin.Default()

	server := &HttpServer{
		nameResolver: nameResolver,
		router:       router,
	}

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/name/:name", server.handleNameQuery)

	return server
}

func (s *HttpServer) Start(address string) error {
	fmt.Printf("Starting HTTP server on %s\n", address)
	return s.router.Run(address)
}

func (s *HttpServer) handleNameQuery(c *gin.Context) {
	name := c.Param("name")
	result, err := s.nameResolver.Resolve(name)

	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
