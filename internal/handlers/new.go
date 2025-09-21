package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterNewRoutes(rg *gin.RouterGroup) {
	rg.GET("/new", func(c *gin.Context) {
		id := uuid.NewString()
		c.Data(http.StatusOK, "text/plain", []byte(id))
	})
}
