package ckan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	baseUrl = "https://data.gov.ro/api/3/action"
)

type Client struct {
	client *http.Client
	url    *url.URL
}

func New() (*Client, error) {
	parsedUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %v", err)
	}
	return &Client{
		client: &http.Client{},
		url:    parsedUrl,
	}, nil
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"__type"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Type)
}

type Response[T any] struct {
	Help    string `json:"help"`
	Success bool   `json:"success"`
	Result  T      `json:"result"`
	Error   *Error `json:"error,omitempty"`
}

func doRequest[T any](ctx context.Context, client *Client, url *url.URL) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	var result Response[T]
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}
	if !result.Success && result.Error != nil {
		return nil, result.Error
	}

	return &result.Result, nil
}
