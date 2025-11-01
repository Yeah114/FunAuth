package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrProxyDisabled 表示未启用代理池（环境变量未配置）。
	ErrProxyDisabled = errors.New("proxy pool disabled")
)

const (
	defaultScheme         = "http"
	defaultRequestTimeout = 5 * time.Second
	defaultClientTimeout  = 30 * time.Second
	defaultDialTimeout    = 10 * time.Second
	defaultIdleTimeout    = 60 * time.Second
)

type config struct {
	endpoint       string
	scheme         string
	requestTimeout time.Duration
	clientTimeout  time.Duration
}

// NewHTTPClient 根据环境变量配置获取代理池中的一个节点并返回配置好的 http.Client。
//
// 依赖的环境变量：
//   - FUNAUTH_PROXY_API_URL: 代理池 HTTP API（必需）
//   - FUNAUTH_PROXY_SCHEME: 代理协议（可选，默认 http）
//   - FUNAUTH_PROXY_REQUEST_TIMEOUT: 请求代理池接口的超时时间（可选，默认 5s）
//   - FUNAUTH_PROXY_CLIENT_TIMEOUT: 使用代理访问业务接口的超时时间（可选，默认 30s）
func NewHTTPClient(ctx context.Context) (*http.Client, error) {
	cfg, err := loadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	proxyURL, err := acquireProxy(ctx, cfg)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout:   defaultDialTimeout,
		KeepAlive: defaultIdleTimeout,
	}

	transport := &http.Transport{
		Proxy:               http.ProxyURL(proxyURL),
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: defaultDialTimeout,
		MaxIdleConns:        32,
		IdleConnTimeout:     defaultIdleTimeout,
		ForceAttemptHTTP2:   true,
	}

	return &http.Client{
		Timeout:   cfg.clientTimeout,
		Transport: transport,
	}, nil
}

func loadConfigFromEnv() (config, error) {
	endpoint := strings.TrimSpace(os.Getenv("FUNAUTH_PROXY_API_URL"))
	if endpoint == "" {
		return config{}, ErrProxyDisabled
	}

	scheme := strings.TrimSpace(os.Getenv("FUNAUTH_PROXY_SCHEME"))
	if scheme == "" {
		scheme = defaultScheme
	}

	requestTimeout, err := parseDurationEnv("FUNAUTH_PROXY_REQUEST_TIMEOUT", defaultRequestTimeout)
	if err != nil {
		return config{}, fmt.Errorf("parse FUNAUTH_PROXY_REQUEST_TIMEOUT: %w", err)
	}

	clientTimeout, err := parseDurationEnv("FUNAUTH_PROXY_CLIENT_TIMEOUT", defaultClientTimeout)
	if err != nil {
		return config{}, fmt.Errorf("parse FUNAUTH_PROXY_CLIENT_TIMEOUT: %w", err)
	}

	return config{
		endpoint:       endpoint,
		scheme:         scheme,
		requestTimeout: requestTimeout,
		clientTimeout:  clientTimeout,
	}, nil
}

func parseDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, err
	}
	return d, nil
}

func acquireProxy(ctx context.Context, cfg config) (*url.URL, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create proxy request: %w", err)
	}

	client := &http.Client{Timeout: cfg.requestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request proxy pool: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("proxy pool status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read proxy response: %w", err)
	}

	proxyURL, err := parseProxyResponse(body, cfg.scheme)
	if err != nil {
		return nil, err
	}

	return proxyURL, nil
}

func parseProxyResponse(body []byte, scheme string) (*url.URL, error) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil, errors.New("proxy pool returned empty body")
	}

	if strings.HasPrefix(trimmed, "{") {
		url, err := parseJSONProxy(body, scheme)
		if err == nil {
			return url, nil
		}
	}

	return parsePlainProxy(trimmed, scheme)
}

func parseJSONProxy(body []byte, scheme string) (*url.URL, error) {
	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, err
	}

	if ret, ok := intFromAny(root["ret"]); ok && ret != 200 && ret != 0 {
		msg := stringFromAny(root["msg"])
		if msg == "" {
			msg = fmt.Sprintf("ret=%d", ret)
		}
		return nil, fmt.Errorf("proxy pool error: %s", msg)
	}

	if candidate := extractRecord(root); candidate != nil {
		return buildProxyURL(scheme, candidate)
	}

	return nil, errors.New("proxy pool response missing proxy data")
}

type proxyRecord struct {
	ip       string
	port     string
	username string
	password string
}

func extractRecord(root map[string]any) *proxyRecord {
	if ip := stringFromAny(root["ip"]); ip != "" {
		return &proxyRecord{
			ip:       ip,
			port:     stringFromAny(root["port"]),
			username: pickAuthField(root, "user", "username"),
			password: pickAuthField(root, "pwd", "password"),
		}
	}

	if data, ok := root["data"]; ok {
		switch v := data.(type) {
		case []any:
			for _, item := range v {
				if rec := mapToRecord(item); rec != nil {
					return rec
				}
			}
		case map[string]any:
			return mapToRecord(v)
		}
	}

	return nil
}

func mapToRecord(v any) *proxyRecord {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	ip := stringFromAny(m["ip"])
	if ip == "" {
		return nil
	}
	return &proxyRecord{
		ip:       ip,
		port:     stringFromAny(m["port"]),
		username: pickAuthField(m, "user", "username"),
		password: pickAuthField(m, "pwd", "password"),
	}
}

func pickAuthField(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if val := stringFromAny(m[key]); val != "" {
			return val
		}
	}
	return ""
}

func parsePlainProxy(trimmed string, scheme string) (*url.URL, error) {
	line := trimmed
	if idx := strings.IndexAny(line, "\r\n"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	if line == "" {
		return nil, errors.New("proxy pool returned blank line")
	}
	if !strings.Contains(line, "://") {
		line = scheme + "://" + line
	}
	parsed, err := url.Parse(line)
	if err != nil {
		return nil, fmt.Errorf("parse proxy url: %w", err)
	}
	if parsed.Scheme == "" {
		parsed.Scheme = scheme
	}
	if parsed.Host == "" {
		return nil, errors.New("proxy url missing host")
	}
	if _, _, err := net.SplitHostPort(parsed.Host); err != nil {
		return nil, fmt.Errorf("proxy url missing port: %w", err)
	}
	return parsed, nil
}

func buildProxyURL(scheme string, rec *proxyRecord) (*url.URL, error) {
	port := strings.TrimSpace(rec.port)
	if port == "" {
		return nil, errors.New("proxy port missing")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("invalid proxy port %q", port)
	}
	host := net.JoinHostPort(strings.TrimSpace(rec.ip), port)
	proxyURL := &url.URL{Scheme: scheme, Host: host}
	if rec.username != "" || rec.password != "" {
		proxyURL.User = url.UserPassword(rec.username, rec.password)
	}
	return proxyURL, nil
}

func intFromAny(v any) (int, bool) {
	switch val := v.(type) {
	case float64:
		return int(val), true
	case json.Number:
		i, err := val.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	case string:
		i, err := strconv.Atoi(val)
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func stringFromAny(v any) string {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case float64:
		return strconv.FormatInt(int64(val), 10)
	case json.Number:
		return val.String()
	default:
		return ""
	}
}
