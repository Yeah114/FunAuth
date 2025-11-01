package auth

import (
	"context"
	"errors"
	"log"

	g79 "github.com/Yeah114/g79client"

	"github.com/Yeah114/FunAuth/internal/proxy"
)

// NewG79Client 根据当前配置创建带代理能力的 g79 客户端。
//
// 若未配置代理池相关环境变量，则自动回退为直连模式。
func NewG79Client(ctx context.Context) (*g79.Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	httpClient, err := proxy.NewHTTPClient(ctx)
	if err != nil {
		if errors.Is(err, proxy.ErrProxyDisabled) {
			log.Printf("[proxy] FUNAUTH_PROXY_API_URL not set, using direct connection")
			return g79.NewClient()
		}
		return nil, err
	}

	log.Printf("[proxy] using rotating proxy client for g79 requests")
	return g79.NewClientWithHTTPClient(httpClient)
}
