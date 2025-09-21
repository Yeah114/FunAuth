package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Yeah114/FunAuth/auth"
	"github.com/Yeah114/g79client"
	"github.com/gin-gonic/gin"
)

func RegisterPhoenixLoginRoute(api *gin.RouterGroup) {
	api.POST("/phoenix/login", func(c *gin.Context) {
		rawAuthorization := c.GetHeader("Authorization")
		bearerToken := strings.TrimPrefix(rawAuthorization, "Bearer ")
		if bearerToken == "" {
			c.JSON(http.StatusOK, LoginResponse{
				SuccessStates: false,
				Message:       Message{Information: "Login: Authorization header missing Bearer token"},
			})
			return
		}

		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, LoginResponse{
				SuccessStates: false,
				Message:       Message{Information: fmt.Sprintf("Login: 绑定请求体时出现问题, 原因是 %v", err)},
			})
			return
		}
		cookieStr := req.FBToken
		if cookieStr == "" {
			cookieStr = fixedCookie
		}

		cli, err := g79client.NewClient()
		if err != nil {
			c.JSON(http.StatusOK, LoginResponse{
				SuccessStates: false,
				Message:       Message{Information: fmt.Sprintf("Login: 初始化客户端时出现问题, 原因是 %v", err)},
			})
			return
		}

		if err := cli.AuthenticateWithCookie(cookieStr); err != nil {
			c.JSON(http.StatusOK, LoginResponse{
				SuccessStates: false,
				Message:       Message{Information: fmt.Sprintf("Login: 使用 Cookie 认证时出现问题, 原因是 %v", err)},
			})
			return
		}

		loginRes, err := auth.Login(c.Request.Context(), cli, auth.LoginParams{
			ServerCode:      req.ServerCode,
			ServerPassword:  req.ServerPassword,
			ClientPublicKey: req.ClientPublicKey,
		})
		if err != nil {
			c.JSON(http.StatusOK, LoginResponse{
				SuccessStates: false,
				Message:       Message{Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", err)},
			})
			return
		}

		enableSkin := true
		var skinInfo SkinInfo
		if enableSkin {
			authSkinInfo, err := auth.GetSkinInfo(cli)
			if err != nil {
				c.JSON(http.StatusOK, LoginResponse{
					SuccessStates: false,
					Message:       Message{Information: fmt.Sprintf("Login: 获取皮肤信息时出现问题, 原因是 %v", err)},
				})
				return
			}
			skinInfo = SkinInfo{
				ItemID:          authSkinInfo.ItemID,
				SkinDownloadURL: authSkinInfo.SkinDownloadURL,
				SkinIsSlim:      authSkinInfo.SkinIsSlim,
			}
		}

		resetSession(bearerToken)
		session := getSessionByBearer(c)
		if session == nil {
			c.JSON(http.StatusOK, LoginResponse{
				SuccessStates: false,
				Message:       Message{Information: fmt.Sprintf("Login: 无效的 Auth Bearer (%s)", rawAuthorization)},
			})
			return
		}
		session.Store(sessionKeyEntityID, loginRes.EntityID)
		session.Store(sessionKeyEngineVersion, loginRes.EngineVersion)
		session.Store(sessionKeyPatchVersion, loginRes.PatchVersion)
		session.Store(sessionKeyUserID, loginRes.UID)

		resp := LoginResponse{
			SuccessStates:  true,
			Message:        Message{Information: "ok"},
			BotLevel:       loginRes.BotLevel,
			BotSkin:        skinInfo,
			BotComponent:   loginRes.BotComponent,
			FBToken:        req.FBToken,
			MasterName:     loginRes.MasterName,
			RentalServerIP: loginRes.IP,
			ChainInfo:      loginRes.ChainInfo,
		}
		c.JSON(http.StatusOK, resp)
	})
}
