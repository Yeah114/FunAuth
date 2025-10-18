package auth

// LoginParams 定义进入服务器验证所需的参数。
type LoginParams struct {
	ServerCode      string
	ServerPassword  string
	ClientPublicKey string
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
}

type SkinInfo struct {
	ItemID          string
	SkinDownloadURL string
	SkinIsSlim      bool
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
