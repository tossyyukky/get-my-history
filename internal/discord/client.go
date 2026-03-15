package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
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

func (c *Client) FetchMessages(ctx context.Context, channelID string, window domain.DigestWindow) ([]domain.Message, error) {
	var (
		before   string
		messages []domain.Message
	)

	for {
		batch, err := c.getChannelMessages(ctx, channelID, before)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}

		stop := false
		for _, item := range batch {
			if item.Type != 0 {
				continue
			}

			ts, err := time.Parse(time.RFC3339, item.Timestamp)
			if err != nil {
				return nil, fmt.Errorf("parse message timestamp %s: %w", item.ID, err)
			}

			if ts.Before(window.Start) {
				stop = true
				break
			}
			if !ts.Before(window.End) {
				continue
			}

			messages = append(messages, domain.Message{
				ID:        item.ID,
				AuthorID:  item.Author.ID,
				Author:    item.Author.Username,
				Content:   buildContent(item),
				Timestamp: ts.UTC(),
				URL:       messageURL(item.GuildID, channelID, item.ID),
			})
		}

		if stop || len(batch) < 100 {
			break
		}

		before = batch[len(batch)-1].ID
	}

	slices.Reverse(messages)
	return messages, nil
}

func (c *Client) SendDigest(ctx context.Context, channelID, mentionUserID, notionURL, summary string) error {
	body := map[string]any{
		"content": buildDigestMessage(mentionUserID, notionURL, summary),
	}
	if mentionUserID != "" {
		body["allowed_mentions"] = map[string]any{
			"parse": []string{},
			"users": []string{mentionUserID},
		}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal discord digest payload: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID),
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("build discord create message request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+c.cfg.DiscordBotToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord create message request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord create message failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	return nil
}

func (c *Client) getChannelMessages(ctx context.Context, channelID, before string) ([]messageResponse, error) {
	query := url.Values{}
	query.Set("limit", "100")
	if before != "" {
		query.Set("before", before)
	}

	endpoint := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages?%s", channelID, query.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build discord get messages request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+c.cfg.DiscordBotToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discord get messages request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord get messages failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var data []messageResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode discord messages response: %w", err)
	}

	return data, nil
}

func buildContent(item messageResponse) string {
	parts := []string{strings.TrimSpace(item.Content)}
	for _, attachment := range item.Attachments {
		if attachment.URL != "" {
			parts = append(parts, attachment.URL)
		}
	}

	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}

	return strings.Join(filtered, "\n")
}

func buildDigestMessage(mentionUserID, notionURL, summary string) string {
	var b strings.Builder
	if mentionUserID != "" {
		b.WriteString("<@")
		b.WriteString(mentionUserID)
		b.WriteString(">\n")
	}
	b.WriteString("今週のDiscordまとめです。\n")
	if notionURL != "" {
		b.WriteString("Notion: ")
		b.WriteString(notionURL)
		b.WriteString("\n\n")
	}
	b.WriteString(summary)
	return b.String()
}

func messageURL(guildID, channelID, messageID string) string {
	if guildID == "" {
		guildID = "@me"
	}

	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, messageID)
}

type messageResponse struct {
	ID          string              `json:"id"`
	GuildID     string              `json:"guild_id"`
	Content     string              `json:"content"`
	Timestamp   string              `json:"timestamp"`
	Type        int                 `json:"type"`
	Author      messageAuthor       `json:"author"`
	Attachments []messageAttachment `json:"attachments"`
}

type messageAuthor struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type messageAttachment struct {
	URL string `json:"url"`
}
