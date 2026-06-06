package auth

import (
	g79 "github.com/Yeah114/g79client"
)

// LoginParams 定义进入服务器验证所需的参数。
type LoginParams struct {
	ServerCode         string
	ServerPassword     string
	ClientPublicKey    string
	AutoUpdateNickname bool
	LoginCCVoice       bool
	NoLogin            bool
}

// LoginResult 为登录/进入服务器后的结果。
type LoginResult struct {
	UID           string
	ChainInfo     string
	IP            string
	BotLevel      int
	MasterName    string
	BotComponent  map[string]*int
	EntityID      string
	EngineVersion string
	PatchVersion  string
	IsPC          bool
	CCVoice       CCVoice

	// G79Client 是经过认证的 G79 客户端(仅租赁服 + 用户中心路径下非 nil)。
	// 上层在 Minecraft 连接建立后应使用此客户端创建 LinkConnection 并发送 GameStart。
	G79Client *g79.Client
	// LocalVitality 是本地 g79client 活力 API，认证成功后始终可用。
	LocalVitality *LocalVitalityAPI
	// GameStartPayload 是租赁服场景下需要发送的 GameStart 负载(仅租赁服场景下非 nil)。
	GameStartPayload map[string]interface{}
}

type SkinInfo struct {
	ItemID          string
	SkinDownloadURL string
	SkinIsSlim      bool
}

type CCVoiceSubUrlInfo struct {
	Http  string
	Https string
}

type CCVoiceSDKConfig struct {
	StaticResUrl CCVoiceSubUrlInfo
	ConfigUrl    CCVoiceSubUrlInfo
}

type CCVoiceInfo struct {
	StreamName       string
	Account          string
	FastReconnection int
	Uid              string
	GameUid          int64
	StatUrl          CCVoiceSubUrlInfo
	HttpKey          string
	QueryUrl         CCVoiceSubUrlInfo
	ChannelType      string
	Game             int
	CheckUrl         CCVoiceSubUrlInfo
	SDKConfigs       CCVoiceSDKConfig
}

type CCVoiceData struct {
	Info       CCVoiceInfo
	StreamName string
	Ts         string
	Sign       string
	Eid        int64
	Streamid   string
	Nodes      []string
}

type CCVoice struct {
	ChannelType string
	Stream      string
	Data        CCVoiceData
}

type TanLobbyLoginParams struct {
	RoomID string
}

type TanLobbyLoginResult struct {
	RoomOwnerID            uint32
	UserUniqueID           uint32
	UserPlayerName         string
	BotLevel               int
	BotComponent           map[string]*int
	RaknetServerAddress    string
	RoomModDisplayName     []string
	RoomModDownloadURL     []string
	RoomModEncryptKey      [][]byte
	SignalingServerAddress string

	RaknetRand      []byte
	RaknetAESRand   []byte
	EncryptKeyBytes []byte
	DecryptKeyBytes []byte

	SignalingSeed   []byte
	SignalingTicket []byte
}

type TanLobbyCreateResult struct {
	UserUniqueID           uint32
	UserPlayerName         string
	RaknetServerAddress    string
	RaknetRand             []byte
	RaknetAESRand          []byte
	EncryptKeyBytes        []byte
	DecryptKeyBytes        []byte
	SignalingServerAddress string
	SignalingSeed          []byte
	SignalingTicket        []byte
}
