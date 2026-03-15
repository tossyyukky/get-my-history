package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tossy-yukky/codex-sample-project/internal/config"
	"github.com/tossy-yukky/codex-sample-project/internal/domain"
)

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Summarize(ctx context.Context, digest domain.WeeklyDigest) (string, error) {
	body := map[string]any{
		"model":        c.cfg.OpenAIModel,
		"instructions": "You summarize weekly Discord activity in Japanese. Keep it concise, accurate, and useful as a reminder. Use plain text only.",
		"input":        buildPrompt(digest),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build openai request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.OpenAIAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai summarize failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var data struct {
		OutputText string `json:"output_text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}
	if strings.TrimSpace(data.OutputText) == "" {
		return "", fmt.Errorf("openai response did not include output_text")
	}

	return strings.TrimSpace(data.OutputText), nil
}

func buildPrompt(digest domain.WeeklyDigest) string {
	jst := time.FixedZone("JST", 9*60*60)
	var b strings.Builder
	fmt.Fprintf(&b, "期間: %s から %s まで (JST)\n", digest.Window.Start.In(jst).Format("2006-01-02 15:04"), digest.Window.End.In(jst).Format("2006-01-02 15:04"))
	fmt.Fprintf(&b, "件数: %d\n\n", digest.MessageCount())
	b.WriteString("以下のDiscord投稿を週次で振り返る要約を日本語で作ってください。\n")
	b.WriteString("出力条件:\n")
	b.WriteString("- 4から8行程度\n")
	b.WriteString("- 重要な話題、決定事項、次に見るべきことを優先\n")
	b.WriteString("- 推測を書かない\n")
	b.WriteString("- 箇条書き中心\n\n")
	if digest.MessageCount() == 0 {
		b.WriteString("投稿はありませんでした。\n")
		return b.String()
	}

	for _, msg := range digest.Messages {
		fmt.Fprintf(&b, "[%s] %s: %s\n", msg.Timestamp.In(jst).Format("2006-01-02 15:04"), msg.Author, msg.Content)
	}

	return b.String()
}
