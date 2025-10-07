package handlers

import (
	"fmt"
	"net/http"

	auth "github.com/Yeah114/FunAuth/auth"
	"github.com/gin-gonic/gin"
)

func RegisterPhoenixTransferStartTypeRoute(api *gin.RouterGroup) {
	api.GET("/phoenix/transfer_start_type", func(c *gin.Context) {
		var q TransferStartTypeQuery
		if err := c.ShouldBindQuery(&q); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		session := getSessionByBearer(c)
		if session == nil {
			c.Status(http.StatusBadRequest)
			return
		}
		userIDRaw, ok := session.Load(sessionKeyUserID)
		if !ok {
			c.Status(http.StatusBadRequest)
			return
		}
		userID, ok := userIDRaw.(string)
		if !ok || userID == "" {
			c.Status(http.StatusBadRequest)
			return
		}
		enc, err := auth.TransferStartType(userID, q.Content)
		if err != nil {
			c.JSON(http.StatusOK, TransferStartTypeResponse{
				Success: false,
				Message: fmt.Sprintf("TransferStartType: %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, TransferStartTypeResponse{Success: true, Message: "ok", Data: enc})
	})
}
