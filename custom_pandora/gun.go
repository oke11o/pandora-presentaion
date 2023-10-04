package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yandex/pandora/core"
	"github.com/yandex/pandora/core/aggregator/netsample"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type GunConfig struct {
	Target    string `validate:"required"`
	Transport TransportConfig
	Sleep     *time.Duration
}
type TransportConfig struct {
	IdleConnTimeout time.Duration
}

func defaultGunConfig() GunConfig {
	return GunConfig{
		Transport: TransportConfig{
			IdleConnTimeout: 1,
		},
	}
}

type Gun struct {
	conf   GunConfig
	aggr   netsample.Aggregator
	deps   core.GunDeps
	client http.Client
}

func (g *Gun) Bind(aggr core.Aggregator, deps core.GunDeps) error {
	g.aggr = netsample.UnwrapAggregator(aggr)
	g.deps = deps
	tr := &http.Transport{
		IdleConnTimeout: g.conf.Transport.IdleConnTimeout,
	}
	g.client = http.Client{Transport: tr}
	return nil
}

func (g *Gun) Shoot(ammo core.Ammo) {
	a, ok := ammo.(*Ammo)
	if !ok {
		g.deps.Log.Error("unexpected ammo type", zap.Any("ammo", ammo))
		return
	}
	sleep := func() {}
	if g.conf.Sleep != nil {
		sleep = func() {
			time.Sleep(*g.conf.Sleep)
		}
	}

	g.shoot(a, sleep)
}

func (g *Gun) shoot(ammo *Ammo, sleep func()) {
	ctx := context.Background()
	authToken, err := g.auth(ctx, ammo.UserID)
	if err != nil {
		g.deps.Log.Error("cant get auth token", zap.Error(err))
	}
	sleep()

	itemIDs, err := g.list(ctx, ammo.UserID, authToken)
	if err != nil {
		g.deps.Log.Error("cant get item list", zap.Error(err))
	}
	sleep()

	for i := 0; i < 3; i++ {
		itemID := itemIDs[rand.Intn(len(itemIDs))]
		err := g.order(ctx, itemID, ammo.UserID, authToken)
		if err != nil {
			g.deps.Log.Error("cant get item list", zap.Error(err))
			return
		}
		sleep()
	}
}

func (g *Gun) auth(ctx context.Context, userID int64) (token string, err error) {
	sample := netsample.Acquire("auth")
	sampleCode := 0
	defer func() {
		sample.SetProtoCode(sampleCode)
		g.aggr.Report(sample)
	}()

	addr := fmt.Sprintf("http://%s/auth", g.conf.Target)
	body := strings.NewReader(fmt.Sprintf(`{"user_id": %d}`, userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, body)
	if err != nil {
		sampleCode = http.StatusBadRequest
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		sampleCode = resp.StatusCode
		return "", fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	result := struct {
		AuthKey string `json:"auth_key"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		sampleCode = http.StatusBadRequest
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.AuthKey, nil
}

func (g *Gun) list(ctx context.Context, id int64, token string) ([]int64, error) {
	sample := netsample.Acquire("list")
	sampleCode := 0
	defer func() {
		sample.SetProtoCode(sampleCode)
		g.aggr.Report(sample)
	}()

	addr := fmt.Sprintf("http://%s/list", g.conf.Target)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		sampleCode = http.StatusBadRequest
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.client.Do(req)
	if err != nil {
		sampleCode = resp.StatusCode
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	result := struct {
		ItemIDs []int64 `json:"items"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		sampleCode = http.StatusBadRequest
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ItemIDs, nil
}

func (g *Gun) order(ctx context.Context, itemID int64, userID int64, token string) error {
	sample := netsample.Acquire("auth")
	sampleCode := 0
	defer func() {
		sample.SetProtoCode(sampleCode)
		g.aggr.Report(sample)
	}()

	addr := fmt.Sprintf("http://%s/order", g.conf.Target)
	body := strings.NewReader(fmt.Sprintf(`{"user_id": %d, "item_id": %d}`, userID, itemID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, body)
	if err != nil {
		sampleCode = http.StatusBadRequest
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.client.Do(req)
	if err != nil {
		sampleCode = resp.StatusCode
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	result := struct {
		Order int64 `json:"order"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		sampleCode = http.StatusBadRequest
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

var _ core.Gun = new(Gun)
