package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestWaitDomainServerIDWithRetriesUntilServerAppears(t *testing.T) {
	calls := 0
	id, err := waitDomainServerIDWith(func() ([]string, error) {
		calls++
		if calls < 3 {
			return nil, nil
		}
		return []string{"", " sid-123 "}, nil
	}, 4, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "sid-123" {
		t.Fatalf("expected trimmed sid, got %q", id)
	}
	if calls != 3 {
		t.Fatalf("expected 3 fetch attempts, got %d", calls)
	}
}

func TestWaitDomainServerIDWithReturnsNotFoundWhenStillEmpty(t *testing.T) {
	_, err := waitDomainServerIDWith(func() ([]string, error) {
		return []string{}, nil
	}, 3, 0)
	if err == nil || !strings.Contains(err.Error(), "加入后未找到山头服务器") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestWaitDomainServerIDWithReturnsLastFetchError(t *testing.T) {
	_, err := waitDomainServerIDWith(func() ([]string, error) {
		return nil, errors.New("boom")
	}, 2, 0)
	if err == nil || !strings.Contains(err.Error(), "GetOtherDomainServers(after join): boom") {
		t.Fatalf("expected wrapped fetch error, got %v", err)
	}
}
