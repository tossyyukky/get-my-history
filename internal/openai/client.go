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
		OutputText string               `json:"output_text"`
		Output     []responseOutputItem `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}
	text := strings.TrimSpace(data.OutputText)
	if text == "" {
		text = strings.TrimSpace(extractOutputText(data.Output))
	}
	if text == "" {
		return "", fmt.Errorf("openai response did not include output_text")
	}

	return text, nil
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

	missingMessages := digest.MessagesWithMissingContent()
	if len(missingMessages) > 0 {
		b.WriteString("\n内容取得失敗の投稿:\n")
		for _, msg := range missingMessages {
			fmt.Fprintf(&b, "- [%s] %s の投稿は本文が取得できませんでした\n", msg.Timestamp.In(jst).Format("2006-01-02 15:04"), msg.Author)
		}
	}

	if len(digest.References) > 0 {
		b.WriteString("\n参照した公開URLの内容:\n")
		for _, ref := range digest.References {
			if ref.Failed() {
				fmt.Fprintf(&b, "- URL: %s\n", ref.URL)
				fmt.Fprintf(&b, "  取得失敗: %s\n", ref.Error)
				continue
			}
			if ref.Title != "" {
				fmt.Fprintf(&b, "- %s\n", ref.Title)
			}
			fmt.Fprintf(&b, "  URL: %s\n", ref.URL)
			fmt.Fprintf(&b, "  抜粋: %s\n", ref.Excerpt)
		}
	}

	return b.String()
}

func extractOutputText(items []responseOutputItem) string {
	var parts []string
	for _, item := range items {
		for _, content := range item.Content {
			if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
				parts = append(parts, strings.TrimSpace(content.Text))
			}
		}
	}

	return strings.Join(parts, "\n")
}

type responseOutputItem struct {
	Content []responseContentItem `json:"content"`
}

type responseContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
