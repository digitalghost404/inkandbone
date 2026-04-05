package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultOllamaURL = "http://localhost:11434"

// OllamaClient calls a local Ollama instance via its OpenAI-compatible API.
// It implements Completer, Responder, and Streamer.
type OllamaClient struct {
	model   string
	baseURL string
	http    *http.Client
}

// NewOllamaClient creates an OllamaClient for the given model using the default
// localhost:11434 base URL. Override with OLLAMA_HOST env var via NewOllamaClientWithURL.
func NewOllamaClient(model string) *OllamaClient {
	return &OllamaClient{model: model, baseURL: defaultOllamaURL, http: &http.Client{}}
}

// NewOllamaClientWithURL is like NewOllamaClient but uses the given base URL (for tests).
func NewOllamaClientWithURL(model, baseURL string) *OllamaClient {
	return &OllamaClient{model: model, baseURL: baseURL, http: &http.Client{}}
}

// Generate implements Completer. Sends a single-turn prompt and returns the response.
func (c *OllamaClient) Generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	return c.chatOnce(ctx, "", []ChatMessage{{Role: "user", Content: prompt}}, maxTokens)
}

// Respond implements Responder. Sends a multi-turn conversation with a system prompt.
func (c *OllamaClient) Respond(ctx context.Context, system string, history []ChatMessage, maxTokens int) (string, error) {
	text, err := c.chatOnce(ctx, system, history, maxTokens)
	if err != nil {
		return "", err
	}
	return stripEmDash(text), nil
}

// StreamRespond implements Streamer. Streams the response as SSE data lines to w.
func (c *OllamaClient) StreamRespond(ctx context.Context, system string, history []ChatMessage, maxTokens int, w http.ResponseWriter) (string, error) {
	msgs := ollamaMessages(system, history)
	body, err := json.Marshal(map[string]any{
		"model":      c.model,
		"messages":   msgs,
		"max_tokens": maxTokens,
		"stream":     true,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned %d: %s", resp.StatusCode, strings.TrimSpace(string(errBody)))
	}

	flusher, canFlush := w.(http.Flusher)
	var fullText strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		data, found := strings.CutPrefix(line, "data: ")
		if !found || data == "[DONE]" {
			continue
		}
		var event struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
			text := stripEmDash(event.Choices[0].Delta.Content)
			fullText.WriteString(text)
			fmt.Fprintf(w, "data: %s\n\n", text) //nolint:errcheck
			if canFlush {
				flusher.Flush()
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fullText.String(), fmt.Errorf("read stream: %w", err)
	}
	return fullText.String(), nil
}

func (c *OllamaClient) chatOnce(ctx context.Context, system string, history []ChatMessage, maxTokens int) (string, error) {
	msgs := ollamaMessages(system, history)
	body, err := json.Marshal(map[string]any{
		"model":      c.model,
		"messages":   msgs,
		"max_tokens": maxTokens,
		"stream":     false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned %d: %s", resp.StatusCode, strings.TrimSpace(string(errBody)))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response from Ollama")
	}
	return result.Choices[0].Message.Content, nil
}

// ollamaMessages converts a system prompt and history into OpenAI-format messages.
func ollamaMessages(system string, history []ChatMessage) []map[string]string {
	var msgs []map[string]string
	if system != "" {
		msgs = append(msgs, map[string]string{"role": "system", "content": system})
	}
	for _, m := range history {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	return msgs
}

// DualOllamaClient routes GM streaming/response calls to one model and all
// automation Generate calls to another. This lets you use a RP-tuned model
// (e.g. hermes3:8b) for narrative prose and a stronger instruction-following
// model (e.g. phi4:14b) for structured JSON tasks.
type DualOllamaClient struct {
	gm   *OllamaClient // handles Respond + StreamRespond
	auto *OllamaClient // handles Generate
}

// NewDualOllamaClient creates a split-model Ollama client.
func NewDualOllamaClient(gmModel, autoModel string) *DualOllamaClient {
	return &DualOllamaClient{
		gm:   NewOllamaClient(gmModel),
		auto: NewOllamaClient(autoModel),
	}
}

func (d *DualOllamaClient) Generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	return d.auto.Generate(ctx, prompt, maxTokens)
}

func (d *DualOllamaClient) Respond(ctx context.Context, system string, history []ChatMessage, maxTokens int) (string, error) {
	return d.gm.Respond(ctx, system, history, maxTokens)
}

func (d *DualOllamaClient) StreamRespond(ctx context.Context, system string, history []ChatMessage, maxTokens int, w http.ResponseWriter) (string, error) {
	return d.gm.StreamRespond(ctx, system, history, maxTokens, w)
}
