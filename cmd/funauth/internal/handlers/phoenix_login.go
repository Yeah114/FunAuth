package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Yeah114/FunAuth/auth"
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

		cli, err := auth.NewG79Client(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusOK, LoginResponse{
				SuccessStates: false,
				Message:       Message{Information: fmt.Sprintf("Login: 初始化客户端时出现问题, 原因是 %v", err)},
			})
			return
		}

		if err := cli.G79AuthenticateWithCookie(cookieStr); err != nil {
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
			LoginCCVoice:    req.LoginCCVoice,
			NoLogin:         req.NoLogin,
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
		session.Store(sessionKeyIsPC, loginRes.IsPC)

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
			CCVoice:        mapCCVoice(loginRes.CCVoice),
		}
		c.JSON(http.StatusOK, resp)
	})
}

func mapCCVoice(ccVoice auth.CCVoice) *CCVoice {
	if ccVoice.ChannelType == "" && ccVoice.Stream == "" && ccVoice.Data.StreamName == "" {
		return nil
	}

	return &CCVoice{
		ChannelType: ccVoice.ChannelType,
		Stream:      ccVoice.Stream,
		Data: CCVoiceData{
			Info: CCVoiceInfo{
				StreamName:       ccVoice.Data.Info.StreamName,
				Account:          ccVoice.Data.Info.Account,
				FastReconnection: ccVoice.Data.Info.FastReconnection,
				Uid:              ccVoice.Data.Info.Uid,
				GameUid:          ccVoice.Data.Info.GameUid,
				StatUrl:          mapCCVoiceSubUrlInfo(ccVoice.Data.Info.StatUrl),
				HttpKey:          ccVoice.Data.Info.HttpKey,
				QueryUrl:         mapCCVoiceSubUrlInfo(ccVoice.Data.Info.QueryUrl),
				ChannelType:      ccVoice.Data.Info.ChannelType,
				Game:             ccVoice.Data.Info.Game,
				CheckUrl:         mapCCVoiceSubUrlInfo(ccVoice.Data.Info.CheckUrl),
				SDKConfigs: CCVoiceSDKConfig{
					StaticResUrl: mapCCVoiceSubUrlInfo(ccVoice.Data.Info.SDKConfigs.StaticResUrl),
					ConfigUrl:    mapCCVoiceSubUrlInfo(ccVoice.Data.Info.SDKConfigs.ConfigUrl),
				},
			},
			StreamName: ccVoice.Data.StreamName,
			Ts:         ccVoice.Data.Ts,
			Sign:       ccVoice.Data.Sign,
			Eid:        ccVoice.Data.Eid,
			Streamid:   ccVoice.Data.Streamid,
			Nodes:      ccVoice.Data.Nodes,
		},
	}
}

func mapCCVoiceSubUrlInfo(info auth.CCVoiceSubUrlInfo) CCVoiceSubUrlInfo {
	return CCVoiceSubUrlInfo{
		Http:  info.Http,
		Https: info.Https,
	}
}
