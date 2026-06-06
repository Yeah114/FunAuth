package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	g79 "github.com/Yeah114/g79client"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// Login
func Login(ctx context.Context, cli *g79.Client, p LoginParams) (LoginResult, error) {
	var result LoginResult
	if cli == nil {
		return result, fmt.Errorf("nil client")
	}

	// 确保用户详情可用，用于昵称与等级
	if cli.UserDetail == nil {
		detail, err := cli.GetUserDetail()
		if err != nil {
			return result, fmt.Errorf("GetUserDetail: %w", err)
		}
		cli.UserDetail = &detail.Entity
	}
	if (cli.UserDetail != nil && cli.UserDetail.Name == "") || p.AutoUpdateNickname {
		name := fmt.Sprintf("AE%09d", random.Intn(1000000000))
		if err := cli.UpdateNickname(name); err != nil {
			return result, fmt.Errorf("UpdateNickname: %w", err)
		}
	}

	// IP
	var ipAddress string
	// ChainInfo
	var chainInfoStr string

	if p.ServerCode == "" {
		return result, fmt.Errorf("server code is empty")
	}
	p.ServerPassword = strings.ReplaceAll(p.ServerPassword, "000000", "")

	if after, ok := strings.CutPrefix(p.ServerCode, "LobbyGame:"); ok && after != "" {
		// 联机大厅
		roomCode := after
		if len(roomCode) != 19 {
			searchResp, err := cli.SearchOnlineLobbyRoomByKeyword(roomCode, 1, 0)
			if err != nil {
				return result, fmt.Errorf("SearchOnlineLobbyRoomByKeyword: %w", err)
			}
			if searchResp.Code != 0 {
				return result, fmt.Errorf("SearchOnlineLobbyRoomByKeyword: %s(%d)", searchResp.Message, searchResp.Code)
			}
			if len(searchResp.Entities) == 0 {
				return result, fmt.Errorf("SearchOnlineLobbyRoomByKeyword: 找不到房间")
			}
			roomCode = searchResp.Entities[0].EntityID.String()
		}

		// 获取房间信息
		roomInfo, err := cli.GetOnlineLobbyRoom(roomCode)
		if err != nil {
			return result, fmt.Errorf("GetOnlineLobbyRoom: %w", err)
		}
		if roomInfo.Code != 0 {
			return result, fmt.Errorf("GetOnlineLobbyRoom: %s(%d)", roomInfo.Message, roomInfo.Code)
		}

		// 购买房间地图
		roomMap, err := cli.PurchaseItem(roomInfo.Entity.ResID.String())
		if err != nil {
			return result, fmt.Errorf("PurchaseItem: %w", err)
		}
		if !(roomMap.Code == 0 || roomMap.Code == 502 || roomMap.Code == 44) {
			return result, fmt.Errorf("PurchaseItem: %s(%d)", roomMap.Message, roomMap.Code)
		}

		// 进入房间（含 HTTP 瞬态错误、501、12022 频率限制重试）
		_, err = enterLobbyRoom(cli, roomCode, p.ServerPassword, func() {
			_, _ = cli.PurchaseItem(roomInfo.Entity.ResID.String())
		})
		if err != nil {
			return result, err
		}

		// 进入房间游戏（含 HTTP 瞬态错误重试）
		gameEnter, err := lobbyGameEnter(cli)
		if err != nil {
			return result, err
		}
		ipAddress = fmt.Sprintf("%s:%d", gameEnter.Entity.ServerHost, gameEnter.Entity.ServerPort.Int64())

		if !p.NoLogin {
			// 获取 ChainInfo
			authv2Data, err := cli.GenerateLobbyGameAuthV2(roomCode, p.ClientPublicKey)
			if err != nil {
				return result, fmt.Errorf("GenerateLobbyGameAuthV2: %w", err)
			}
			chainInfo, err := cli.SendAuthV2Request(authv2Data)
			if err != nil {
				return result, fmt.Errorf("SendAuthV2Request: %w", err)
			}
			chainInfoStr = string(chainInfo)
		}
	} else if after, ok := strings.CutPrefix(p.ServerCode, "PCLobbyGame:"); ok && after != "" {
		// PC联机大厅
		if cli.ReleaseJSON.ApiGatewayUrl != cli.X19ReleaseJSON.ApiGatewayGrayURL {
			newCli, err := g79.NewClient()
			if err != nil {
				return result, fmt.Errorf("NewClient: %w", err)
			}
			for {
				time.Sleep(time.Second * 10)
				err = newCli.X19AuthenticateWithCookie(cli.Cookie)
				if err == nil {
					break
				}
				if strings.Contains(err.Error(), "操作过于频繁，请稍后重试") {
					continue
				} else {
					return result, fmt.Errorf("X19AuthenticateWithCookie: %w", err)
				}
			}
			cli = newCli
		}
		roomCode := after
		if len(roomCode) != 19 {
			searchResp, err := cli.SearchOnlineLobbyRoomByKeyword(roomCode, 1, 0)
			if err != nil {
				return result, fmt.Errorf("SearchOnlineLobbyRoomByKeyword: %w", err)
			}
			if searchResp.Code != 0 {
				return result, fmt.Errorf("SearchOnlineLobbyRoomByKeyword: %s(%d)", searchResp.Message, searchResp.Code)
			}
			if len(searchResp.Entities) == 0 {
				return result, fmt.Errorf("SearchOnlineLobbyRoomByKeyword: 找不到房间")
			}
			roomCode = searchResp.Entities[0].EntityID.String()
		}

		// 获取房间信息
		roomInfo, err := cli.GetOnlineLobbyRoom(roomCode)
		if err != nil {
			return result, fmt.Errorf("GetOnlineLobbyRoom: %w", err)
		}
		if roomInfo.Code != 0 {
			return result, fmt.Errorf("GetOnlineLobbyRoom: %s(%d)", roomInfo.Message, roomInfo.Code)
		}

		// 购买房间地图
		roomMap, err := cli.UserItemPurchase(roomInfo.Entity.ResID.String())
		if err != nil {
			return result, fmt.Errorf("UserItemPurchase: %w", err)
		}
		if !(roomMap.Code == 0 || roomMap.Code == 502 || roomMap.Code == 44) {
			return result, fmt.Errorf("UserItemPurchase: %s(%d)", roomMap.Message, roomMap.Code)
		}

		// 进入房间（含 HTTP 瞬态错误、501、12022 频率限制重试）
		_, err = enterLobbyRoom(cli, roomCode, p.ServerPassword, func() {
			_, _ = cli.UserItemPurchase(roomInfo.Entity.ResID.String())
		})
		if err != nil {
			return result, err
		}

		// 进入房间游戏（含 HTTP 瞬态错误重试）
		gameEnter, err := lobbyGameEnter(cli)
		if err != nil {
			return result, err
		}
		ipAddress = fmt.Sprintf("%s:%d", gameEnter.Entity.ServerHost, gameEnter.Entity.ServerPort.Int64())

		if !p.NoLogin {
			// 获取 ChainInfo
			authv2Data, err := cli.GeneratePCLobbyGameAuthV2(roomInfo.Entity.ResID.String(), p.ClientPublicKey)
			if err != nil {
				return result, fmt.Errorf("GeneratePCLobbyGameAuthV2: %w", err)
			}
			chainInfo, err := cli.SendAuthV2Request(authv2Data)
			if err != nil {
				return result, fmt.Errorf("SendAuthV2Request: %w", err)
			}
			chainInfoStr = string(chainInfo)
		}
		result.IsPC = true
	} else if after, ok := strings.CutPrefix(p.ServerCode, "NetworkGame:"); ok && after != "" {
		// 网络游戏
		gameCode := after

		// 获取网络游戏服务器地址
		serverAddress, err := cli.GetPeGameServerAddress(gameCode)
		if err != nil {
			return result, fmt.Errorf("GetPeGameServerAddress: %w", err)
		}
		if serverAddress.Code != 0 {
			return result, fmt.Errorf("GetPeGameServerAddress: %s(%d)", serverAddress.Message, serverAddress.Code)
		}
		ipAddress = fmt.Sprintf("%s:%d", serverAddress.Entity.IP, serverAddress.Entity.Port.Int64())

		if !p.NoLogin {
			// 生成网络游戏认证v2数据
			authv2Data, err := cli.GenerateNetworkGameAuthV2(gameCode, p.ClientPublicKey)
			if err != nil {
				return result, fmt.Errorf("GenerateNetworkGameAuthV2: %w", err)
			}
			chainInfo, err := cli.SendAuthV2Request(authv2Data)
			if err != nil {
				return result, fmt.Errorf("SendAuthV2Request: %w", err)
			}
			chainInfoStr = string(chainInfo)
		}
	} else if p.ServerCode == "MainCity" {
		// 网易主城
		_ = cli.LeaveEnteredGame()
		_, _ = cli.LeaveMainCity()
		mainCity, err := cli.EnterMainCity()
		if err != nil {
			return result, fmt.Errorf("EnterMainCity: %w", err)
		}
		if mainCity.Code != 0 {
			return result, fmt.Errorf("EnterMainCity: %s(%d)", mainCity.Message, mainCity.Code)
		}
		ipAddress = fmt.Sprintf("%s:%d", mainCity.Entity.ServerHost, mainCity.Entity.ServerPort)
		if !p.NoLogin {
			authv2Data, err := cli.GenerateLobbyGameAuthV2(fmt.Sprintf("%d", mainCity.Entity.CityNo), p.ClientPublicKey)
			if err != nil {
				return result, fmt.Errorf("GenerateLobbyGameAuthV2: %w", err)
			}
			chainInfo, err := cli.SendAuthV2Request(authv2Data)
			if err != nil {
				return result, fmt.Errorf("SendAuthV2Request: %w", err)
			}
			chainInfoStr = string(chainInfo)
		}
	} else if after, ok := strings.CutPrefix(p.ServerCode, "DomainGame:"); ok && after != "" {
		// 我的山头
		inviteCode := after

		// 清理已存在的其他山头服务器（避免冲突）
		resp, err := cli.GetOtherDomainServers()
		if err != nil {
			return result, fmt.Errorf("GetOtherDomainServers: %w", err)
		}
		for _, server := range resp.Entities {
			if _, delErr := cli.DeleteOtherDomainServer(server.Sid); delErr != nil {
				return result, fmt.Errorf("DeleteOtherDomainServer: %w", delErr)
			}
		}

		// 通过邀请码加入山头服务器
		inviteResp, err := cli.JoinDomainServerWithInviteCode(inviteCode)
		if err != nil {
			return result, fmt.Errorf("JoinDomainServerWithInviteCode: %w", err)
		}
		if inviteResp.Code != 0 {
			return result, fmt.Errorf("JoinDomainServerWithInviteCode: %s(%d)", inviteResp.Message, inviteResp.Code)
		}

		// 获取加入后的山头服务器ID
		serverID, err := waitDomainServerID(cli)
		if err != nil {
			return result, err
		}

		// 请求进入山头服务器（获取IP和端口）
		enterResp, err := cli.RequestEnterDomainServer(serverID)
		if err != nil {
			return result, fmt.Errorf("RequestEnterDomainServer(sid=%s): %w", serverID, err)
		}
		if enterResp.Code != 0 {
			return result, fmt.Errorf("RequestEnterDomainServer: %s(%d)", enterResp.Message, enterResp.Code)
		}
		ipAddress = fmt.Sprintf("%s:%d", enterResp.Entity.ServerHost, enterResp.Entity.ServerPort.Int64())

		if p.LoginCCVoice {
			ccVoice, err := getCCVoiceLoginInfo(cli, serverID)
			if err != nil {
				return result, err
			}
			result.CCVoice = ccVoice
		}

		if !p.NoLogin {
			// 生成DomainGame认证数据并获取ChainInfo
			authv2Data, err := cli.GenerateDomainGameAuthV2(serverID, p.ClientPublicKey)
			if err != nil {
				return result, fmt.Errorf("GenerateDomainGameAuthV2: %w", err)
			}
			chainInfo, err := cli.SendAuthV2Request(authv2Data)
			if err != nil {
				return result, fmt.Errorf("SendAuthV2Request: %w", err)
			}
			chainInfoStr = string(chainInfo)
		}

		// 请求离开山头服务器
		leaveResp, err := cli.RequestLeaveDomainServer(serverID)
		if err != nil {
			return result, fmt.Errorf("RequestLeaveDomainServer(sid=%s): %w", serverID, err)
		}
		if leaveResp.Code != 0 {
			return result, fmt.Errorf("RequestLeaveDomainServer: %s(%d)", leaveResp.Message, leaveResp.Code)
		}

		// 删除山头服务器
		if _, delErr := cli.DeleteOtherDomainServer(serverID); delErr != nil {
			return result, fmt.Errorf("DeleteOtherDomainServer(after auth-v2): %w", delErr)
		}
	} else if after, ok := strings.CutPrefix(p.ServerCode, "PCDomainGame:"); ok && after != "" {
		// PC我的山头
		inviteCode := after

		// 清理已存在的其他山头服务器（避免冲突）
		resp, err := cli.GetOtherDomainServers()
		if err != nil {
			return result, fmt.Errorf("GetOtherDomainServers: %w", err)
		}
		for _, server := range resp.Entities {
			if _, delErr := cli.DeleteOtherDomainServer(server.Sid); delErr != nil {
				return result, fmt.Errorf("DeleteOtherDomainServer: %w", delErr)
			}
		}

		// 通过邀请码加入山头服务器
		inviteResp, err := cli.JoinDomainServerWithInviteCode(inviteCode)
		if err != nil {
			return result, fmt.Errorf("JoinDomainServerWithInviteCode: %w", err)
		}
		if inviteResp.Code != 0 {
			return result, fmt.Errorf("JoinDomainServerWithInviteCode: %s(%d)", inviteResp.Message, inviteResp.Code)
		}

		// 获取加入后的山头服务器ID
		serverID, err := waitDomainServerID(cli)
		if err != nil {
			return result, err
		}

		// 请求进入山头服务器（获取IP和端口）
		enterResp, err := cli.RequestEnterDomainServer(serverID)
		if err != nil {
			return result, fmt.Errorf("RequestEnterDomainServer(sid=%s): %w", serverID, err)
		}
		if enterResp.Code != 0 {
			return result, fmt.Errorf("RequestEnterDomainServer: %s(%d)", enterResp.Message, enterResp.Code)
		}
		ipAddress = fmt.Sprintf("%s:%d", enterResp.Entity.ServerHost, enterResp.Entity.ServerPort.Int64())

		if p.LoginCCVoice {
			ccVoice, err := getCCVoiceLoginInfo(cli, serverID)
			if err != nil {
				return result, err
			}
			result.CCVoice = ccVoice
		}

		if !p.NoLogin {
			// 生成PCDomainGame认证数据并获取ChainInfo
			authv2Data, err := cli.GeneratePCDomainGameAuthV2(serverID, p.ClientPublicKey)
			if err != nil {
				return result, fmt.Errorf("GeneratePCDomainGameAuthV2: %w", err)
			}
			chainInfo, err := cli.SendAuthV2Request(authv2Data)
			if err != nil {
				return result, fmt.Errorf("SendAuthV2Request: %w", err)
			}
			chainInfoStr = string(chainInfo)
		}

		// 请求离开山头服务器
		leaveResp, err := cli.RequestLeaveDomainServer(serverID)
		if err != nil {
			return result, fmt.Errorf("RequestLeaveDomainServer(sid=%s): %w", serverID, err)
		}
		if leaveResp.Code != 0 {
			return result, fmt.Errorf("RequestLeaveDomainServer: %s(%d)", leaveResp.Message, leaveResp.Code)
		}

		// 删除山头服务器
		if _, delErr := cli.DeleteOtherDomainServer(serverID); delErr != nil {
			return result, fmt.Errorf("DeleteOtherDomainServer(after auth-v2): %w", delErr)
		}
		result.IsPC = true
	} else {
		// 租赁服
		serverCode := p.ServerCode

		// 搜索租赁服
		searchResp, err := cli.SearchRentalServerByName(serverCode)
		if err != nil {
			return result, fmt.Errorf("SearchRentalServerByName: %w", err)
		}
		if searchResp.Code != 0 {
			return result, fmt.Errorf("SearchRentalServerByName: %s(%d)", searchResp.Message, searchResp.Code)
		}
		if len(searchResp.Entities) == 0 {
			return result, fmt.Errorf("SearchRentalServerByName: 找不到服务器")
		}
		serverEntity := searchResp.Entities[0]
		serverID := serverEntity.EntityID

		ownerID := strings.TrimSpace(serverEntity.OwnerID.String())
		var ownerName string
		if ownerID != "" {
			ownerInfo, err := cli.GetUserElseDetailMany([]string{ownerID})
			if err != nil {
				return result, fmt.Errorf("GetUserElseDetailMany(owner_id=%s): %w", ownerID, err)
			}
			if ownerInfo.Code != 0 {
				return result, fmt.Errorf("GetUserElseDetailMany: %s(%d)", ownerInfo.Message, ownerInfo.Code)
			}
			if len(ownerInfo.Entities) == 0 {
				return result, fmt.Errorf("GetUserElseDetailMany: 未返回服主信息")
			}
			ownerName = ownerInfo.Entities[0].Nickname
		}
		if ownerName == "" {
			if ownerID != "" {
				ownerName = ownerID
			} else {
				ownerName = serverCode
			}
		}

		// 进入租赁服世界
		enterResp, err := cli.EnterRentalServerWorld(serverID.String(), p.ServerPassword)
		if err != nil {
			return result, fmt.Errorf("EnterRentalServerWorld: %w", err)
		}
		if enterResp.Code != 0 {
			return result, fmt.Errorf("EnterRentalServerWorld: %s(%d)", enterResp.Message, enterResp.Code)
		}
		ipAddress = fmt.Sprintf("%s:%d", enterResp.Entity.McserverHost, enterResp.Entity.McserverPort.Int64())

		// 构建 GameStart 负载，但不在此处发送。
		// 将 G79 客户端和负载传递给上层，在 Minecraft 连接建立后再发送。
		gameInfo := map[string]interface{}{
			"min_level": serverEntity.MinLevel.Int64(),
			"room_name": serverCode,
			"gameType":  "RentalGame",
			"res_name":  serverCode,
			"ownerName": ownerName,
			"ownerId":   ownerID,
			"id":        serverEntity.EntityID.String(),
		}
		gameInfoJSON, err := json.Marshal(gameInfo)
		if err != nil {
			return result, fmt.Errorf("marshal rental game info: %w", err)
		}
		result.G79Client = cli
		result.GameStartPayload = map[string]interface{}{
			"game_info":    string(gameInfoJSON),
			"strict_mode":  true,
			"game_type":    10,
			"is_free_play": false,
			"game_id":      serverEntity.EntityID.String(),
			"play_iids":    []string{},
		}

		if !p.NoLogin {
			// 获取 ChainInfo
			authv2Data, err := cli.GenerateRentalGameAuthV2(serverID.String(), p.ClientPublicKey)
			if err != nil {
				return result, fmt.Errorf("GenerateRentalGameAuthV2: %w", err)
			}
			chainInfo, err := cli.SendAuthV2Request(authv2Data)
			if err != nil {
				return result, fmt.Errorf("SendAuthV2Request: %w", err)
			}
			chainInfoStr = string(chainInfo)
		}
	}

	result.UID = cli.UserID
	result.EntityID = cli.UserDetail.EntityID
	result.MasterName = cli.UserDetail.Name
	if result.MasterName == "" {
		result.MasterName = cli.UserID
	}
	result.ChainInfo = chainInfoStr
	result.IP = ipAddress
	result.BotLevel = int(cli.UserDetail.Level.Int64())
	result.EngineVersion = cli.EngineVersion
	result.PatchVersion = cli.G79LatestVersion
	if result.G79Client == nil {
		result.G79Client = cli
	}
	if result.LocalVitality == nil {
		localVitality, err := NewLocalVitalityAPI(result.G79Client)
		if err != nil {
			return result, fmt.Errorf("NewLocalVitalityAPI: %w", err)
		}
		result.LocalVitality = localVitality
	}

	return result, nil
}

func getCCVoiceLoginInfo(cli *g79.Client, serverID string) (CCVoice, error) {
	var ccVoice CCVoice
	voiceResp, err := cli.GetCCVoiceLoginInfo(serverID)
	if err != nil {
		return ccVoice, fmt.Errorf("GetCCVoiceLoginInfo(sid=%s): %w", serverID, err)
	}
	if voiceResp.Code != 0 {
		return ccVoice, fmt.Errorf("GetCCVoiceLoginInfo: %s(%d)", voiceResp.Message, voiceResp.Code)
	}

	data := voiceResp.Entity.DataEntity
	info := data.InfoData
	ccVoice = CCVoice{
		ChannelType: voiceResp.Entity.ChannelType,
		Stream:      voiceResp.Entity.Stream,
		Data: CCVoiceData{
			Info: CCVoiceInfo{
				StreamName:       info.StreamName,
				Account:          info.Account,
				FastReconnection: info.FastReconnection,
				Uid:              info.Uid,
				GameUid:          info.GameUid,
				StatUrl:          mapCCVoiceSubUrlInfo(info.StatUrl),
				HttpKey:          info.HttpKey,
				QueryUrl:         mapCCVoiceSubUrlInfo(info.QueryUrl),
				ChannelType:      info.ChannelType,
				Game:             info.Game,
				CheckUrl:         mapCCVoiceSubUrlInfo(info.CheckUrl),
				SDKConfigs: CCVoiceSDKConfig{
					StaticResUrl: mapCCVoiceSubUrlInfo(info.SDKConfigs.StaticResUrl),
					ConfigUrl:    mapCCVoiceSubUrlInfo(info.SDKConfigs.ConfigUrl),
				},
			},
			StreamName: data.StreamName,
			Ts:         data.Ts,
			Sign:       data.Sign,
			Eid:        data.Eid,
			Streamid:   data.Streamid,
			Nodes:      data.Nodes,
		},
	}

	return ccVoice, nil
}

func mapCCVoiceSubUrlInfo(info g79.CCVoiceSubUrlInfo) CCVoiceSubUrlInfo {
	return CCVoiceSubUrlInfo{
		Http:  info.Http,
		Https: info.Https,
	}
}
