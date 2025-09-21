package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Yeah114/FunAuth/auth"
	"github.com/Yeah114/g79client"
	"github.com/gin-gonic/gin"
)

func RegisterPhoenixTanLobbyCreateRoute(api *gin.RouterGroup) {
	api.POST("/phoenix/tan_lobby_create", func(c *gin.Context) {
		rawAuthorization := c.GetHeader("Authorization")
		authorization := strings.TrimPrefix(rawAuthorization, "Bearer ")
		if authorization == "" {
			c.JSON(http.StatusOK, TanLobbyCreateResponse{Success: false, ErrorInfo: "TanLobbyCreate: Authorization header missing Bearer token"})
			return
		}

		var req TanLobbyCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, TanLobbyCreateResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyCreate: 绑定请求体时出现问题, 原因是 %v", err)})
			return
		}
		cookieStr := req.FBToken
		if cookieStr == "" {
			cookieStr = fixedCookie
		}

		cli, err := g79client.NewClient()
		if err != nil {
			c.JSON(http.StatusOK, TanLobbyCreateResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyCreate: 初始化客户端时出现问题, 原因是 %v", err)})
			return
		}

		if err := cli.AuthenticateWithCookie(cookieStr); err != nil {
			c.JSON(http.StatusOK, TanLobbyCreateResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyCreate: 使用 Cookie 认证时出现问题, 原因是 %v", err)})
			return
		}

		createRes, err := auth.TanLobbyCreate(c.Request.Context(), cli)
		if err != nil {
			c.JSON(http.StatusOK, TanLobbyCreateResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyCreate: %v", err)})
			return
		}

		c.JSON(http.StatusOK, TanLobbyCreateResponse{
			Success:                true,
			ErrorInfo:              "",
			UserUniqueID:           createRes.UserUniqueID,
			UserPlayerName:         createRes.UserPlayerName,
			RaknetServerAddress:    createRes.RaknetServerAddress,
			RaknetRand:             createRes.RaknetRand,
			RaknetAESRand:          createRes.RaknetAESRand,
			EncryptKeyBytes:        createRes.EncryptKeyBytes,
			DecryptKeyBytes:        createRes.DecryptKeyBytes,
			SignalingServerAddress: createRes.SignalingServerAddress,
			SignalingSeed:          createRes.SignalingSeed,
			SignalingTicket:        createRes.SignalingTicket,
		})
	})
}
