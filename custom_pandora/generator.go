package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/yandex/pandora/core"
	"github.com/yandex/pandora/core/aggregator/netsample"
	coreimport "github.com/yandex/pandora/core/import"
	"github.com/yandex/pandora/core/register"
	"go.uber.org/zap"
)

func Import(fs afero.Fs) {
	coreimport.RegisterCustomJSONProvider("custom_provider",
		func() core.Ammo {
			return &Payload{}
		},
	)
	register.Gun("custom_generator", func(conf GeneratorConfig) core.Gun {
		return &Generator{
			conf: conf,
		}
	}, defaultConfig)
}

type Payload struct {
	UserID   int64  `json:"user_id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type GeneratorConfig struct {
	Target    string `validate:"required"`
	Transport TransportConfig
	Sleep     time.Duration
	ReqSleep  time.Duration
}
type TransportConfig struct {
	IdleConnTimeout time.Duration
}

func defaultConfig() GeneratorConfig {
	return GeneratorConfig{
		Transport: TransportConfig{
			IdleConnTimeout: 1,
		},
		ReqSleep: 100 * time.Millisecond,
	}
}

type Generator struct {
	conf   GeneratorConfig
	aggr   netsample.Aggregator
	deps   core.GunDeps
	client http.Client
}

func (g *Generator) Bind(aggr core.Aggregator, deps core.GunDeps) error {
	g.aggr = netsample.UnwrapAggregator(aggr)
	g.deps = deps
	tr := &http.Transport{
		IdleConnTimeout: g.conf.Transport.IdleConnTimeout,
	}
	g.client = http.Client{Transport: tr}
	return nil
}

func (g *Generator) Shoot(payload core.Ammo) {
	a, ok := payload.(*Payload)
	if !ok {
		g.deps.Log.Error("unexpected payload type", zap.Any("payload", payload))
		return
	}
	g.shoot(a)
}

func (g *Generator) shoot(payload *Payload) {
	ctx := context.Background()
	authToken, err := g.auth(ctx, payload.UserID)
	if err != nil {
		g.deps.Log.Error("cant get auth token", zap.Error(err))
		return
	}
	time.Sleep(g.conf.Sleep)

	itemIDs, err := g.list(ctx, payload.UserID, authToken)
	if err != nil {
		g.deps.Log.Error("cant get item list", zap.Error(err))
		return
	}
	time.Sleep(g.conf.Sleep)

	for i := 0; i < 3; i++ {
		itemID := itemIDs[rand.Intn(len(itemIDs))]
		err := g.order(ctx, itemID, payload.UserID, authToken)
		if err != nil {
			g.deps.Log.Error("cant get item list", zap.Error(err))
			return
		}
		time.Sleep(g.conf.Sleep)
	}
}

func (g *Generator) auth(ctx context.Context, userID int64) (token string, err error) {
	sample := netsample.Acquire("auth")
	sampleCode := 0
	defer func() {
		sample.SetProtoCode(sampleCode)
		g.aggr.Report(sample)
	}()

	addr := fmt.Sprintf("http://%s/auth?sleep=%d", g.conf.Target, g.conf.Sleep.Milliseconds())
	body := strings.NewReader(fmt.Sprintf(`{"user_id": %d}`, userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, body)
	if err != nil {
		sampleCode = http.StatusBadRequest
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	sampleCode = resp.StatusCode
	if err != nil {
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

func (g *Generator) list(ctx context.Context, id int64, token string) ([]int64, error) {
	sample := netsample.Acquire("list")
	sampleCode := 0
	defer func() {
		sample.SetProtoCode(sampleCode)
		g.aggr.Report(sample)
	}()

	addr := fmt.Sprintf("http://%s/list?sleep=%d", g.conf.Target, g.conf.Sleep.Milliseconds())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		sampleCode = http.StatusBadRequest
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.client.Do(req)
	sampleCode = resp.StatusCode
	if err != nil {
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

func (g *Generator) order(ctx context.Context, itemID int64, userID int64, token string) error {
	sample := netsample.Acquire("order")
	sampleCode := 0
	defer func() {
		sample.SetProtoCode(sampleCode)
		g.aggr.Report(sample)
	}()

	addr := fmt.Sprintf("http://%s/order?sleep=%d", g.conf.Target, g.conf.Sleep.Milliseconds())
	body := strings.NewReader(fmt.Sprintf(`{"user_id": %d, "item_id": %d}`, userID, itemID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, body)
	if err != nil {
		sampleCode = http.StatusBadRequest
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.client.Do(req)
	sampleCode = resp.StatusCode
	if err != nil {
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

var _ core.Gun = new(Generator)
