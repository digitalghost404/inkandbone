package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultURL       = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
	model            = "claude-haiku-4-5-20251001"
)

// Completer generates text from a prompt. Implemented by *Client; nil means AI is disabled.
type Completer interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// ChatMessage is a single turn in a conversation.
type ChatMessage struct {
	Role    string
	Content string
}

// Responder generates a reply from a system prompt and conversation history.
type Responder interface {
	Respond(ctx context.Context, system string, history []ChatMessage) (string, error)
}

// Client calls the Anthropic Messages API over plain HTTP.
type Client struct {
	apiKey string
	url    string
	http   *http.Client
}

// NewClient returns a Client using the production Anthropic API URL.
func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, url: defaultURL, http: &http.Client{}}
}

// NewClientWithURL returns a Client using a custom URL (for tests).
func NewClientWithURL(apiKey, url string) *Client {
	return &Client{apiKey: apiKey, url: url, http: &http.Client{}}
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 1024,
		"messages":   []map[string]any{{"role": "user", "content": prompt}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return "", fmt.Errorf("anthropic API returned %d", resp.StatusCode)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic")
	}
	return result.Content[0].Text, nil
}

func (c *Client) Respond(ctx context.Context, system string, history []ChatMessage) (string, error) {
	msgs := make([]map[string]any, len(history))
	for i, m := range history {
		msgs[i] = map[string]any{"role": m.Role, "content": m.Content}
	}
	body, err := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 2048,
		"system":     system,
		"messages":   msgs,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return "", fmt.Errorf("anthropic API returned %d", resp.StatusCode)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic")
	}
	return result.Content[0].Text, nil
}
