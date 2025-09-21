package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Yeah114/FunAuth/auth"
	"github.com/Yeah114/g79client"
	"github.com/gin-gonic/gin"
)

func RegisterPhoenixTanLobbyLoginRoute(api *gin.RouterGroup) {
	api.POST("/phoenix/tan_lobby_login", func(c *gin.Context) {
		rawAuthorization := c.GetHeader("Authorization")
		authorization := strings.TrimPrefix(rawAuthorization, "Bearer ")
		if authorization == "" {
			c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: "TanLobbyLogin: Authorization header missing Bearer token"})
			return
		}

		var req TanLobbyLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: 绑定请求体时出现问题, 原因是 %v", err)})
			return
		}
		cookieStr := req.FBToken
		if cookieStr == "" {
			cookieStr = fixedCookie
		}

		cli, err := g79client.NewClient()
		if err != nil {
			c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: 初始化客户端时出现问题, 原因是 %v", err)})
			return
		}

		if err := cli.AuthenticateWithCookie(cookieStr); err != nil {
			c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: 使用 Cookie 认证时出现问题, 原因是 %v", err)})
			return
		}

		loginRes, err := auth.TanLobbyLogin(c.Request.Context(), cli, auth.TanLobbyLoginParams{
			RoomID: req.RoomID,
		})
		if err != nil {
			c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err)})
			return
		}

		enableSkin := true
		var skinInfo SkinInfo
		if enableSkin {
			authSkinInfo, err := auth.GetSkinInfo(cli)
			if err != nil {
				c.JSON(http.StatusOK, TanLobbyLoginResponse{Success: false, ErrorInfo: fmt.Sprintf("TanLobbyLogin: 获取皮肤信息时出现问题, 原因是 %v", err)})
				return
			}
			skinInfo = SkinInfo{
				ItemID:          authSkinInfo.ItemID,
				SkinDownloadURL: authSkinInfo.SkinDownloadURL,
				SkinIsSlim:      authSkinInfo.SkinIsSlim,
			}
		}

		botLevel := 0
		if cli.UserDetail != nil {
			botLevel = int(cli.UserDetail.Level.Int64())
		}
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:                true,
			ErrorInfo:              "",
			UserUniqueID:           loginRes.UserUniqueID,
			UserPlayerName:         loginRes.UserPlayerName,
			BotLevel:               botLevel,
			BotSkin:                skinInfo,
			BotComponent:           loginRes.BotComponent,
			RoomOwnerID:            loginRes.RoomOwnerID,
			RoomModDisplayName:     loginRes.RoomModDisplayName,
			RoomModDownloadURL:     loginRes.RoomModDownloadURL,
			RoomModEncryptKey:      loginRes.RoomModEncryptKey,
			RaknetServerAddress:    loginRes.RaknetServerAddress,
			RaknetRand:             loginRes.RaknetRand,
			RaknetAESRand:          loginRes.RaknetAESRand,
			EncryptKeyBytes:        loginRes.EncryptKeyBytes,
			DecryptKeyBytes:        loginRes.DecryptKeyBytes,
			SignalingServerAddress: loginRes.SignalingServerAddress,
			SignalingSeed:          loginRes.SignalingSeed,
			SignalingTicket:        loginRes.SignalingTicket,
		})
	})
}
