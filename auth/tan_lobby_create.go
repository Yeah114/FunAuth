package auth

import (
	"context"
	cryptoRand "crypto/rand"
	"fmt"
	"math/rand"
	"time"

	"github.com/Yeah114/g79client"
	"github.com/Yeah114/g79client/utils"
)

// TanLobbyCreate sets up transfer information required to host a tan lobby room.
func TanLobbyCreate(ctx context.Context, cli *g79client.Client) (TanLobbyCreateResult, error) {
	var result TanLobbyCreateResult

	if cli == nil {
		return result, fmt.Errorf("TanLobbyCreate: nil client")
	}

	if cli.UserToken == "" {
		return result, fmt.Errorf("TanLobbyCreate: missing user token")
	}

	if cli.UserDetail == nil {
		detail, err := cli.GetUserDetail()
		if err != nil {
			return result, fmt.Errorf("TanLobbyCreate: GetUserDetail: %w", err)
		}
		cli.UserDetail = &detail.Entity
	}

	raknetAddr, signalingAddr, err := selectTransferServer(cli)
	if err != nil {
		return result, fmt.Errorf("TanLobbyCreate: %w", err)
	}

	encryptedToken := utils.GetEncryptedToken(cli.UserToken)

	raknetRand := make([]byte, 16)
	if _, err = cryptoRand.Read(raknetRand); err != nil {
		return result, fmt.Errorf("TanLobbyCreate: rand read: %w", err)
	}

	raknetAESRand, err := utils.AesECBEncrypt(raknetRand, encryptedToken)
	if err != nil {
		return result, fmt.Errorf("TanLobbyCreate: aes encrypt: %w", err)
	}
	if len(raknetAESRand) >= 16 {
		raknetAESRand = raknetAESRand[:16]
	}

	encryptKeyBytes := append(append(make([]byte, 0, len(encryptedToken)+len(raknetRand)), encryptedToken...), raknetRand...)
	decryptKeyBytes := append(append(make([]byte, 0, len(encryptedToken)+len(raknetRand)), raknetRand...), encryptedToken...)

	signalingSeed := make([]byte, 16)
	if _, err = cryptoRand.Read(signalingSeed); err != nil {
		return result, fmt.Errorf("TanLobbyCreate: rand read: %w", err)
	}

	signalingTicket, err := utils.AesECBEncrypt(signalingSeed, []byte(cli.UserToken))
	if err != nil {
		return result, fmt.Errorf("TanLobbyCreate: aes encrypt: %w", err)
	}
	if len(signalingTicket) >= 16 {
		signalingTicket = signalingTicket[:16]
	}

	uid, err := cli.GetUserIDInt()
	if err != nil {
		return result, fmt.Errorf("TanLobbyCreate: parse user id: %w", err)
	}

	playerName := cli.UserID
	if cli.UserDetail != nil && cli.UserDetail.Name != "" {
		playerName = cli.UserDetail.Name
	}

	result.UserUniqueID = uint32(uid)
	result.UserPlayerName = playerName
	result.RaknetServerAddress = raknetAddr
	result.RaknetRand = raknetRand
	result.RaknetAESRand = raknetAESRand
	result.EncryptKeyBytes = encryptKeyBytes
	result.DecryptKeyBytes = decryptKeyBytes
	result.SignalingServerAddress = signalingAddr
	result.SignalingSeed = signalingSeed
	result.SignalingTicket = signalingTicket

	return result, nil
}

type transferServerEntry struct {
	Status         int    `json:"status"`
	ServerIP       string `json:"ip"`
	SignalWebPort  int    `json:"SignalWebPort"`
	WebsocketPorts []int  `json:"ports"`
}

func selectTransferServer(cli *g79client.Client) (string, string, error) {
	servers, err := g79client.GetGlobalG79TransferServers()
	if err != nil {
		return "", "", fmt.Errorf("SelectTransferServer: %w", err)
	}

	var list []transferServerEntry
	for _, s := range servers {
		list = append(list, transferServerEntry{
			Status:         int(s.Status.Int64()),
			ServerIP:       s.IP,
			SignalWebPort:  int(s.SignalWebPort.Int64()),
			WebsocketPorts: s.Ports,
		})
	}
	available := make([]transferServerEntry, 0, len(list))
	for _, entry := range list {
		if len(entry.WebsocketPorts) == 0 || entry.ServerIP == "" || entry.SignalWebPort == 0 {
			continue
		}
		available = append(available, entry)
	}
	if len(available) == 0 {
		return "", "", fmt.Errorf("SelectTransferServer: no available server")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	selected := available[rng.Intn(len(available))]
	port := selected.WebsocketPorts[rng.Intn(len(selected.WebsocketPorts))]

	raknetAddr := fmt.Sprintf("%s:%d", selected.ServerIP, port)
	signalingAddr := fmt.Sprintf("%s:%d", selected.ServerIP, selected.SignalWebPort)

	return raknetAddr, signalingAddr, nil
}
