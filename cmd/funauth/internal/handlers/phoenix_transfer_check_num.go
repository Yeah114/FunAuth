package handlers

import (
	"fmt"
	"net/http"

	auth "github.com/Yeah114/FunAuth/auth"
	"github.com/gin-gonic/gin"
)

func RegisterPhoenixTransferCheckNumRoute(api *gin.RouterGroup) {
	api.POST("/phoenix/transfer_check_num", func(c *gin.Context) {
		var req TransferCheckNumRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		session := getSessionByBearer(c)
		if session == nil {
			c.Status(http.StatusBadRequest)
			return
		}
		engineVersionRaw, _ := session.Load(sessionKeyEngineVersion)
		patchVersionRaw, _ := session.Load(sessionKeyPatchVersion)
		engineVersionStr, _ := engineVersionRaw.(string)
		patchVersionStr, _ := patchVersionRaw.(string)

		value, err := auth.TransferCheckNum(
			c.Request.Context(),
			req.Data,
			engineVersionStr,
			patchVersionStr,
		)
		if err != nil {
			c.JSON(http.StatusOK, TransferCheckNumResponse{
				Success: false,
				Message: fmt.Sprintf("TransferCheckNum: %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, TransferCheckNumResponse{Success: true, Message: "ok", Value: value})
	})
}
