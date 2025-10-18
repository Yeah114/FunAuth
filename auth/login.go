package auth

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	g79 "github.com/Yeah114/g79client"
	//link "github.com/Yeah114/g79client/service/link_connection"
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
	if cli.UserDetail != nil && cli.UserDetail.Name == "" {
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

		// 进入房间
		var enterResp *g79.OnlineLobbyRoomEnterResponse
		maxRetries := 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			enterResp, err = cli.EnterOnlineLobbyRoom(roomCode, p.ServerPassword)
			if err != nil {
				return result, fmt.Errorf("EnterOnlineLobbyRoom: %w", err)
			}
			if enterResp.Code != 501 {
				break
			}
			if attempt < maxRetries {
				_, _ = cli.PurchaseItem(roomInfo.Entity.ResID.String())
				time.Sleep(500 * time.Millisecond)
			}
		}
		if enterResp.Code == 501 {
			return result, fmt.Errorf("EnterOnlineLobbyRoom: %s(%d)", enterResp.Message, enterResp.Code)
		}
		if enterResp.Code != 0 {
			return result, fmt.Errorf("EnterOnlineLobbyRoom: %s(%d)", enterResp.Message, enterResp.Code)
		}

		// 进入房间游戏
		gameEnter, err := cli.OnlineLobbyGameEnter()
		if err != nil {
			return result, fmt.Errorf("OnlineLobbyGameEnter: %w", err)
		}
		if gameEnter.Code != 0 {
			return result, fmt.Errorf("OnlineLobbyGameEnter: %s(%d)", gameEnter.Message, gameEnter.Code)
		}
		ipAddress = fmt.Sprintf("%s:%d", gameEnter.Entity.ServerHost, gameEnter.Entity.ServerPort.Int64())

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
	} else if after, ok := strings.CutPrefix(p.ServerCode, "PCLobbyGame:"); ok && after != "" {
		// PC联机大厅
		newCli, err := g79.NewClient()
		if err != nil {
			return result, fmt.Errorf("NewClient: %w", err)
		}
		for {
			time.Sleep(time.Second)
			err = newCli.X19AuthenticateWithCookie(cli.Cookie)
			if err == nil {
				break
			}
			if strings.Contains(err.Error(), "操作过于频繁，请稍后重试") {
				continue
			}
		}
		cli = newCli
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
		

		// 进入房间
		var enterResp *g79.OnlineLobbyRoomEnterResponse
		maxRetries := 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			enterResp, err = cli.EnterOnlineLobbyRoom(roomCode, p.ServerPassword)
			if err != nil {
				return result, fmt.Errorf("EnterOnlineLobbyRoom: %w", err)
			}
			if enterResp.Code != 501 {
				break
			}
			if attempt < maxRetries {
				_, _ = cli.UserItemPurchase(roomInfo.Entity.ResID.String())
				time.Sleep(500 * time.Millisecond)
			}
		}
		if enterResp.Code == 501 {
			return result, fmt.Errorf("EnterOnlineLobbyRoom: %s(%d)", enterResp.Message, enterResp.Code)
		}
		if enterResp.Code != 0 {
			return result, fmt.Errorf("EnterOnlineLobbyRoom: %s(%d)", enterResp.Message, enterResp.Code)
		}

		// 进入房间游戏
		gameEnter, err := cli.OnlineLobbyGameEnter()
		if err != nil {
			return result, fmt.Errorf("OnlineLobbyGameEnter: %w", err)
		}
		if gameEnter.Code != 0 {
			return result, fmt.Errorf("OnlineLobbyGameEnter: %s(%d)", gameEnter.Message, gameEnter.Code)
		}
		ipAddress = fmt.Sprintf("%s:%d", gameEnter.Entity.ServerHost, gameEnter.Entity.ServerPort.Int64())

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
		authv2Data, err := cli.GenerateLobbyGameAuthV2(fmt.Sprintf("%d", mainCity.Entity.CityNo), p.ClientPublicKey)
		if err != nil {
			return result, fmt.Errorf("GenerateLobbyGameAuthV2: %w", err)
		}
		chainInfo, err := cli.SendAuthV2Request(authv2Data)
		if err != nil {
			return result, fmt.Errorf("SendAuthV2Request: %w", err)
		}
		chainInfoStr = string(chainInfo)
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
		serverID := searchResp.Entities[0].EntityID

		// 进入租赁服世界
		enterResp, err := cli.EnterRentalServerWorld(serverID.String(), p.ServerPassword)
		if err != nil {
			return result, fmt.Errorf("EnterRentalServerWorld: %w", err)
		}
		if enterResp.Code != 0 {
			return result, fmt.Errorf("EnterRentalServerWorld: %s(%d)", enterResp.Message, enterResp.Code)
		}
		ipAddress = fmt.Sprintf("%s:%d", enterResp.Entity.McserverHost, enterResp.Entity.McserverPort.Int64())

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
	/*
	service, err := link.NewLinkConnectionService(cli)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := service.Dial(ctx)
	if err != nil {
		return result, err
	}
	defer conn.Close()
	if err := conn.SendGameStart(nil); err != nil {
		return result, err
	}
	if err := conn.SendGameStop(nil); err != nil {
		return result, err
	}
	*/
	return result, nil
}
