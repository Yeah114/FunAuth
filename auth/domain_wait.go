package auth

import (
	"fmt"
	"strings"
	"time"

	g79 "github.com/Yeah114/g79client"
)

const (
	domainServerLookupAttempts = 6
	domainServerLookupDelay    = 500 * time.Millisecond
)

type domainServerIDFetcher func() ([]string, error)

func waitDomainServerID(cli *g79.Client) (string, error) {
	return waitDomainServerIDWith(func() ([]string, error) {
		return listDomainServerIDs(cli)
	}, domainServerLookupAttempts, domainServerLookupDelay)
}

func waitDomainServerIDWith(fetch domainServerIDFetcher, attempts int, delay time.Duration) (string, error) {
	if attempts <= 0 {
		attempts = domainServerLookupAttempts
	}
	if delay < 0 {
		delay = 0
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		ids, err := fetch()
		if err != nil {
			lastErr = err
		} else {
			for _, id := range ids {
				if trimmed := strings.TrimSpace(id); trimmed != "" {
					return trimmed, nil
				}
			}
		}
		if attempt < attempts-1 && delay > 0 {
			time.Sleep(delay)
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("GetOtherDomainServers(after join): %w", lastErr)
	}
	return "", fmt.Errorf("GetOtherDomainServers: 加入后未找到山头服务器")
}

func listDomainServerIDs(cli *g79.Client) ([]string, error) {
	resp, err := cli.GetOtherDomainServers()
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(resp.Entities))
	for _, server := range resp.Entities {
		ids = append(ids, server.Sid)
	}
	return ids, nil
}
