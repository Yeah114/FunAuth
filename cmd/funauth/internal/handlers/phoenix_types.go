package handlers

type LoginRequest struct {
	FBToken         string `json:"login_token,omitempty"`
	UserName        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	ServerCode      string `json:"server_code"`
	ServerPassword  string `json:"server_passcode"`
	ClientPublicKey string `json:"client_public_key"`
	LoginCCVoice    bool   `json:"login_cc_voice,omitempty"`
	NoLogin         bool   `json:"no_login,omitempty"`
}

type SkinInfo struct {
	ItemID          string `json:"entity_id"`
	SkinDownloadURL string `json:"res_url"`
	SkinIsSlim      bool   `json:"is_slim"`
}

type Message struct {
	Information string `json:"message,omitempty"`
	Translation int    `json:"translation,omitempty"`
}

type CCVoiceSubUrlInfo struct {
	Http  string `json:"http"`
	Https string `json:"https"`
}

type CCVoiceSDKConfig struct {
	StaticResUrl CCVoiceSubUrlInfo `json:"static_res_url"`
	ConfigUrl    CCVoiceSubUrlInfo `json:"config_url"`
}

type CCVoiceInfo struct {
	StreamName       string            `json:"stream_name"`
	Account          string            `json:"account"`
	FastReconnection int               `json:"fast_reconnection"`
	Uid              string            `json:"uid"`
	GameUid          int64             `json:"game_uid"`
	StatUrl          CCVoiceSubUrlInfo `json:"stat_url"`
	HttpKey          string            `json:"httpkey"`
	QueryUrl         CCVoiceSubUrlInfo `json:"query_url"`
	ChannelType      string            `json:"channel_type"`
	Game             int               `json:"game"`
	CheckUrl         CCVoiceSubUrlInfo `json:"check_url"`
	SDKConfigs       CCVoiceSDKConfig  `json:"sdk_configs"`
}

type CCVoiceData struct {
	Info       CCVoiceInfo `json:"info"`
	StreamName string      `json:"stream_name"`
	Ts         string      `json:"ts"`
	Sign       string      `json:"sign"`
	Eid        int64       `json:"eid"`
	Streamid   string      `json:"streamid"`
	Nodes      []string    `json:"nodes"`
}

type CCVoice struct {
	ChannelType string      `json:"channel_type"`
	Stream      string      `json:"stream"`
	Data        CCVoiceData `json:"data"`
}

type LoginResponse struct {
	SuccessStates bool   `json:"success"`
	ServerMessage string `json:"server_msg,omitempty"`
	Message
	BotLevel       int             `json:"growth_level"`
	BotSkin        SkinInfo        `json:"skin_info,omitempty"`
	BotComponent   map[string]*int `json:"outfit_info,omitempty"`
	FBToken        string          `json:"token"`
	MasterName     string          `json:"respond_to,omitempty"`
	RentalServerIP string          `json:"ip_address"`
	ChainInfo      string          `json:"chainInfo"`
	CCVoice        *CCVoice        `json:"cc_voice,omitempty"`
}

type TransferCheckNumRequest struct {
	Data          string `json:"data"`
	EngineVersion string `json:"engine_version,omitempty"`
	PatchVersion  string `json:"patch_version,omitempty"`
	IsPC          *bool  `json:"is_pc,omitempty"`
}

type TransferCheckNumResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

type TransferStartTypeQuery struct {
	Content string `form:"content"`
}

type TransferStartTypeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// TanLobbyTransferServersResponse ..
type TanLobbyTransferServersResponse struct {
	Success          bool     `json:"success"`
	ErrorInfo        string   `json:"error_info"`
	RaknetServers    []string `json:"raknet_servers"`
	WebsocketServers []string `json:"websocket_servers"`
}

// TanLobbyLoginRequest ..
type TanLobbyLoginRequest struct {
	FBToken string `json:"login_token"`
	RoomID  string `json:"room_id"`
}

// TanLobbyLoginResponse ..
type TanLobbyLoginResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`

	UserUniqueID   uint32          `json:"user_unique_id"`
	UserPlayerName string          `json:"user_player_name"`
	BotLevel       int             `json:"growth_level"`
	BotSkin        SkinInfo        `json:"skin_info"`
	BotComponent   map[string]*int `json:"outfit_info,omitempty"`

	RoomOwnerID        uint32   `json:"room_owner_id"`
	RoomModDisplayName []string `json:"room_mod_display_name,omitempty"`
	RoomModDownloadURL []string `json:"room_mod_download_url,omitempty"`
	RoomModEncryptKey  [][]byte `json:"room_mod_encrypt_key,omitempty"`

	RaknetServerAddress    string `json:"raknet_server_address"`
	RaknetRand             []byte `json:"raknet_rand"`
	RaknetAESRand          []byte `json:"raknet_aes_rand"`
	EncryptKeyBytes        []byte `json:"encrypt_key_bytes"`
	DecryptKeyBytes        []byte `json:"decrypt_key_bytes"`
	SignalingServerAddress string `json:"signaling_server_address"`
	SignalingSeed          []byte `json:"signaling_seed"`
	SignalingTicket        []byte `json:"signaling_ticket"`
}

type TanLobbyCreateRequest struct {
	FBToken string `json:"login_token"`
}

type TanLobbyCreateResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`

	UserUniqueID           uint32 `json:"user_unique_id"`
	UserPlayerName         string `json:"user_player_name"`
	RaknetServerAddress    string `json:"raknet_server_address"`
	RaknetRand             []byte `json:"raknet_rand"`
	RaknetAESRand          []byte `json:"raknet_aes_rand"`
	EncryptKeyBytes        []byte `json:"encrypt_key_bytes"`
	DecryptKeyBytes        []byte `json:"decrypt_key_bytes"`
	SignalingServerAddress string `json:"signaling_server_address"`
	SignalingSeed          []byte `json:"signaling_seed"`
	SignalingTicket        []byte `json:"signaling_ticket"`
}
