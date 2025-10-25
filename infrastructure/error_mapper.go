package infrastructure

import (
	"bond/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	if domainErr, ok := err.(types.DomainError); ok {
		c.JSON(ErrorToStatusCode(domainErr.Code), gin.H{"error": domainErr.Error()})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

func ErrorToStatusCode(code types.ErrorCode) int {
	switch code {
	case types.ErrorCodeNameRequired:
		return http.StatusBadRequest
	case types.ErrorCodeNameNotFound:
		return http.StatusNotFound
	case types.ErrorCodeRoutingNotFound:
		return http.StatusNotFound
	case types.ErrorCodeStoreFailure:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
