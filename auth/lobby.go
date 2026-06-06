package auth

import (
	"fmt"
	"strings"
	"time"

	g79 "github.com/Yeah114/g79client"
)

const (
	lobbyEnterMaxRetries = 5
	lobbyRateLimitCode   = 12022
)

// isTransientError 判断是否为 HTTP 层瞬态错误（EOF、超时等），这类错误值得重试。
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "TLS handshake") ||
		strings.Contains(msg, "broken pipe")
}

// enterLobbyRoom 封装 EnterOnlineLobbyRoom 的重试逻辑，覆盖：
//   - HTTP 层瞬态错误（EOF/超时等）
//   - 错误码 501（需重新购买地图）
//   - 错误码 12022（网易频率限制）
func enterLobbyRoom(
	cli *g79.Client,
	roomCode, password string,
	repurchase func(),
) (*g79.OnlineLobbyRoomEnterResponse, error) {
	var lastErr error
	for attempt := 1; attempt <= lobbyEnterMaxRetries; attempt++ {
		resp, err := cli.EnterOnlineLobbyRoom(roomCode, password)
		if err != nil {
			if isTransientError(err) && attempt < lobbyEnterMaxRetries {
				lastErr = err
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return nil, fmt.Errorf("EnterOnlineLobbyRoom: %w", err)
		}
		switch resp.Code {
		case 0:
			return resp, nil
		case 501:
			if attempt < lobbyEnterMaxRetries {
				repurchase()
				time.Sleep(500 * time.Millisecond)
				continue
			}
		case lobbyRateLimitCode:
			if attempt < lobbyEnterMaxRetries {
				time.Sleep(time.Duration(3+attempt) * time.Second)
				continue
			}
		default:
			return nil, fmt.Errorf("EnterOnlineLobbyRoom: %s(%d)", resp.Message, resp.Code)
		}
		lastErr = fmt.Errorf("EnterOnlineLobbyRoom: %s(%d)", resp.Message, resp.Code)
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("EnterOnlineLobbyRoom: 重试耗尽")
}

// lobbyGameEnter 封装 OnlineLobbyGameEnter 的重试逻辑，覆盖 HTTP 层瞬态错误。
func lobbyGameEnter(cli *g79.Client) (*g79.OnlineLobbyGameEnterResponse, error) {
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := cli.OnlineLobbyGameEnter()
		if err != nil {
			if isTransientError(err) && attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return nil, fmt.Errorf("OnlineLobbyGameEnter: %w", err)
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("OnlineLobbyGameEnter: %s(%d)", resp.Message, resp.Code)
		}
		return resp, nil
	}
	return nil, fmt.Errorf("OnlineLobbyGameEnter: 重试耗尽")
}
