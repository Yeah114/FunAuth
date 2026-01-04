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

		// 1. 优先从请求体获取参数，获取不到则从 Session 加载
		// 处理 EngineVersion
		engineVersionStr := req.EngineVersion
		if engineVersionStr == "" {
			if session := getSessionByBearer(c); session != nil {
				if raw, ok := session.Load(sessionKeyEngineVersion); ok {
					engineVersionStr, _ = raw.(string)
				}
			}
		}

		// 处理 PatchVersion
		patchVersionStr := req.PatchVersion
		if patchVersionStr == "" {
			if session := getSessionByBearer(c); session != nil {
				if raw, ok := session.Load(sessionKeyPatchVersion); ok {
					patchVersionStr, _ = raw.(string)
				}
			}
		}

		// 处理 IsPC（用指针区分"未传"和"传false"）
		var isPC bool
		if req.IsPC != nil {
			isPC = *req.IsPC
		} else {
			if session := getSessionByBearer(c); session != nil {
				if raw, ok := session.Load(sessionKeyIsPC); ok {
					isPC, _ = raw.(bool)
				}
			}
		}

		// 2. 调用 auth 方法（参数已优先取请求体的值）
		value, err := auth.TransferCheckNum(
			c.Request.Context(),
			isPC,
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

		c.JSON(http.StatusOK, TransferCheckNumResponse{
			Success: true,
			Message: "ok",
			Value:   value,
		})
	})
}
