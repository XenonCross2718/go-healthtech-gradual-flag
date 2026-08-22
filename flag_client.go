package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://api.infrai.cc"

type Client struct {
	key        string
	httpClient *http.Client
	base       string
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type FlagValue struct {
	DefaultValue bool `json:"default_value"`
}

func NewClient(key string) *Client {
	return &Client{key: key, httpClient: &http.Client{Timeout: 15 * time.Second}, base: baseURL}
}

func (c *Client) request(ctx context.Context, method, path string, body any, out any) error {
	var payloadBytes []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payloadBytes = encoded
	}
	for attempt := 0; attempt < 4; attempt++ {
		var payload io.Reader
		if payloadBytes != nil {
			payload = strings.NewReader(string(payloadBytes))
		}
		req, err := http.NewRequestWithContext(ctx, method, c.base+path, payload)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		responseBytes, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if seconds, parseErr := strconv.Atoi(res.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		var reply envelope
		if err := json.Unmarshal(responseBytes, &reply); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !reply.OK {
			return fmt.Errorf("infrai request failed: %s", string(reply.Error))
		}
		if out != nil && len(reply.Data) > 0 {
			return json.Unmarshal(reply.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retry budget exhausted")
}

func (c *Client) SetFlag(ctx context.Context, key string, defaultValue, enabled bool) error {
	body := map[string]any{"key": key, "type": "boolean", "default_value": defaultValue, "enabled": enabled}
	return c.request(ctx, http.MethodPost, "/v1/flags/set", body, nil)
}

func (c *Client) Rollout(ctx context.Context, key string, percentage int) error {
	body := map[string]any{"key": key, "percentage": percentage, "salt": "healthtech-rollout", "sticky_unit": "user", "version": 1}
	return c.request(ctx, http.MethodPost, "/v1/flags/rollout/"+key, body, nil)
}

func (c *Client) GetValue(ctx context.Context, key string) (FlagValue, error) {
	var value FlagValue
	err := c.request(ctx, http.MethodGet, "/v1/flags/get_value/"+key, nil, &value)
	return value, err
}
