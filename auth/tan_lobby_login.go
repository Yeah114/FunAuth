package auth

import (
	"context"
	cryptoRand "crypto/rand"
	"fmt"
	"strconv"
	"strings"

	"github.com/Yeah114/g79client"
	"github.com/Yeah114/g79client/utils"
)

func TanLobbyLogin(ctx context.Context, cli *g79client.Client, p TanLobbyLoginParams) (TanLobbyLoginResult, error) {
	var result TanLobbyLoginResult

	roomInfo, err := cli.GetTransferRoomWithName(p.RoomID)
	if err != nil {
		return result, fmt.Errorf("get transfer room with name: %w", err)
	}
	if roomInfo.Code != 0 {
		return result, fmt.Errorf("get transfer room with name: %s(%d)", roomInfo.Message, roomInfo.Code)
	}
	if len(roomInfo.List) == 0 {
		return result, fmt.Errorf("room not found")
	}

	if cli.UserDetail == nil {
		if detail, err := cli.GetUserDetail(); err == nil {
			cli.UserDetail = &detail.Entity
		}
	}

	encryptedToken := utils.GetEncryptedToken(cli.UserToken)
	raknetRand := make([]byte, 16)
	_, err = cryptoRand.Read(raknetRand)
	if err != nil {
		return result, fmt.Errorf("rand read: %w", err)
	}
	raknetAESRand, err := utils.AesECBEncrypt(raknetRand, encryptedToken)
	if err != nil {
		return result, fmt.Errorf("aes encrypt: %w", err)
	}
	encryptKeyBytes := append(encryptedToken, raknetRand...)
	decryptKeyBytes := append(raknetRand, encryptedToken...)

	seed := make([]byte, 16)
	_, err = cryptoRand.Read(seed)
	if err != nil {
		return result, fmt.Errorf("rand read: %w", err)
	}

	ticket, err := utils.AesECBEncrypt(seed, []byte(cli.UserToken))
	if err != nil {
		return result, fmt.Errorf("aes encrypt: %w", err)
	}

	target := roomInfo.List[0]
	for _, candidate := range roomInfo.List {
		if candidate.RoomUniqueID == p.RoomID || candidate.RID.String() == p.RoomID {
			target = candidate
			break
		}
	}

	result.RoomOwnerID = uint32(target.HID.Int64())
	if cli.UserDetail != nil {
		result.UserPlayerName = cli.UserDetail.Name
	}
	userUniqueID, err := strconv.ParseInt(cli.UserID, 10, 64)
	if err != nil {
		return result, fmt.Errorf("parse int: %w", err)
	}
	result.UserUniqueID = uint32(userUniqueID)
	if cli.UserDetail != nil {
		result.BotLevel = int(cli.UserDetail.Level.Int64())
	}
	result.BotComponent = nil

	itemIDs := make([]string, 0, len(target.ItemIDs))
	for _, rawID := range target.ItemIDs {
		id := strings.TrimSpace(rawID.String())
		if id == "" || id == "0" {
			continue
		}
		itemIDs = append(itemIDs, id)
	}

	if len(itemIDs) > 0 {
		for _, itemID := range itemIDs {
			info, err := cli.GetDownloadInfo(itemID)
			if err != nil {
				return result, fmt.Errorf("get download info: %w", err)
			}
			result.RoomModDisplayName = append(result.RoomModDisplayName, itemID)
			result.RoomModDownloadURL = append(result.RoomModDownloadURL, info.Entity.ResURL)
			result.RoomModEncryptKey = append(result.RoomModEncryptKey, nil)
		}
	}

	roomTransferServerID := int(target.SRV.Int64())
	if roomTransferServerID != 0 {
		servers, err := g79client.GetGlobalG79TransferServers()
		if err != nil {
			return result, fmt.Errorf("get transfer servers: %w", err)
		}
		for _, entry := range servers {
			if int(entry.ID.Int64()) != roomTransferServerID {
				continue
			}
			if len(entry.Ports) > 0 {
				result.RaknetServerAddress = fmt.Sprintf("%s:%d", entry.IP, entry.Ports[0])
			}
			signalPort := entry.SignalWebPort.Int64()
			if signalPort > 0 {
				result.SignalingServerAddress = fmt.Sprintf("%s:%d", entry.IP, signalPort)
			}
			break
		}
	}

	if result.RaknetServerAddress == "" || result.SignalingServerAddress == "" {
		return result, fmt.Errorf("resolve transfer server address failed")
	}
	result.RaknetRand = raknetRand
	result.RaknetAESRand = raknetAESRand
	result.SignalingSeed = seed
	result.SignalingTicket = ticket
	result.EncryptKeyBytes = encryptKeyBytes
	result.DecryptKeyBytes = decryptKeyBytes

	return result, nil
}
