package notion

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
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) CreateWeeklyPage(ctx context.Context, digest domain.WeeklyDigest) (string, error) {
	body := map[string]any{
		"parent": map[string]any{
			"page_id": c.cfg.NotionParentPageID,
		},
		"properties": map[string]any{
			"title": map[string]any{
				"title": []map[string]any{
					{
						"type": "text",
						"text": map[string]any{
							"content": pageTitle(digest.Window),
						},
					},
				},
			},
		},
		"children": buildChildren(digest),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal notion create page payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.notion.com/v1/pages", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build notion create page request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.NotionToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2026-03-11")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("notion create page request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("notion create page failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var data struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode notion create page response: %w", err)
	}

	return data.URL, nil
}

func pageTitle(window domain.DigestWindow) string {
	jst := time.FixedZone("JST", 9*60*60)
	start := window.Start.In(jst)
	end := window.End.In(jst).Add(-time.Nanosecond)
	return fmt.Sprintf("Discord Weekly Digest %s - %s", start.Format("2006-01-02"), end.Format("2006-01-02"))
}

func buildChildren(digest domain.WeeklyDigest) []map[string]any {
	jst := time.FixedZone("JST", 9*60*60)
	children := []map[string]any{
		headingBlock("AI Summary"),
		paragraphBlock(digest.Summary),
		headingBlock("Messages"),
		paragraphBlock(fmt.Sprintf(
			"Collected %d messages from %s to %s (JST).",
			digest.MessageCount(),
			digest.Window.Start.In(jst).Format("2006-01-02 15:04"),
			digest.Window.End.In(jst).Format("2006-01-02 15:04"),
		)),
	}

	if len(digest.Messages) == 0 {
		children = append(children, paragraphBlock("No messages were found for this period."))
		return children
	}

	for _, msg := range digest.Messages {
		line := fmt.Sprintf("[%s] %s: %s", msg.Timestamp.In(jst).Format("2006-01-02 15:04"), msg.Author, msg.Content)
		if msg.URL != "" {
			line += "\n" + msg.URL
		}
		children = append(children, bulletedListItemBlock(truncate(line, 1800)))
	}

	if len(digest.References) > 0 {
		children = append(children, headingBlock("Referenced URLs"))
		for _, ref := range digest.References {
			line := ref.URL
			if ref.Title != "" {
				line = ref.Title + "\n" + line
			}
			if ref.Excerpt != "" {
				line += "\n" + ref.Excerpt
			}
			children = append(children, bulletedListItemBlock(truncate(line, 1800)))
		}
	}

	return children
}

func headingBlock(content string) map[string]any {
	return map[string]any{
		"object": "block",
		"type":   "heading_2",
		"heading_2": map[string]any{
			"rich_text": richText(content),
		},
	}
}

func paragraphBlock(content string) map[string]any {
	return map[string]any{
		"object": "block",
		"type":   "paragraph",
		"paragraph": map[string]any{
			"rich_text": richText(content),
		},
	}
}

func bulletedListItemBlock(content string) map[string]any {
	return map[string]any{
		"object": "block",
		"type":   "bulleted_list_item",
		"bulleted_list_item": map[string]any{
			"rich_text": richText(content),
		},
	}
}

func richText(content string) []map[string]any {
	return []map[string]any{
		{
			"type": "text",
			"text": map[string]any{
				"content": content,
			},
		},
	}
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}

	return value[:limit-3] + "..."
}
