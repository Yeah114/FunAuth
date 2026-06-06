package auth

import (
	"context"
	"fmt"
	"time"

	g79 "github.com/Yeah114/g79client"
)

const DefaultVitalityCurrencyInterval = 5 * time.Minute

type LocalVitalityRequester interface {
	GetCurrencyOnline() (*g79.GetCurrencyOnlineResponse, error)
	GetDailyGrowth() (*g79.DailyGrowthResponse, error)
}

type LocalVitalityAPI struct {
	requester LocalVitalityRequester
}

func NewLocalVitalityAPI(requester LocalVitalityRequester) (*LocalVitalityAPI, error) {
	if requester == nil {
		return nil, fmt.Errorf("nil vitality requester")
	}
	return &LocalVitalityAPI{requester: requester}, nil
}

func (a *LocalVitalityAPI) GetCurrencyOnline() (*g79.GetCurrencyOnlineResponse, error) {
	if a == nil || a.requester == nil {
		return nil, fmt.Errorf("nil local vitality api")
	}
	return a.requester.GetCurrencyOnline()
}

func (a *LocalVitalityAPI) GetDailyGrowth() (*g79.DailyGrowthResponse, error) {
	if a == nil || a.requester == nil {
		return nil, fmt.Errorf("nil local vitality api")
	}
	return a.requester.GetDailyGrowth()
}

func (a *LocalVitalityAPI) MaintainCurrencyOnline(
	ctx context.Context,
	interval time.Duration,
	onError func(error),
) {
	if interval <= 0 {
		interval = DefaultVitalityCurrencyInterval
	}
	report := func(err error) {
		if err != nil && onError != nil {
			onError(err)
		}
	}

	_, err := a.GetCurrencyOnline()
	report(err)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := a.GetCurrencyOnline()
			report(err)
		}
	}
}
