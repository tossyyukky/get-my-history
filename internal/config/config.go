package config

import (
	"errors"
	"fmt"
	"os"
	"sort"
)

type Config struct {
	DiscordBotToken           string
	DiscordSourceChannelID    string
	DiscordNotificationChanID string
	DiscordMentionUserID      string
	NotionToken               string
	NotionParentPageID        string
	OpenAIAPIKey              string
	OpenAIModel               string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		DiscordBotToken:           os.Getenv("DISCORD_BOT_TOKEN"),
		DiscordSourceChannelID:    os.Getenv("DISCORD_SOURCE_CHANNEL_ID"),
		DiscordNotificationChanID: os.Getenv("DISCORD_NOTIFICATION_CHANNEL_ID"),
		DiscordMentionUserID:      os.Getenv("DISCORD_MENTION_USER_ID"),
		NotionToken:               os.Getenv("NOTION_TOKEN"),
		NotionParentPageID:        os.Getenv("NOTION_PARENT_PAGE_ID"),
		OpenAIAPIKey:              os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:               getenvDefault("OPENAI_MODEL", "gpt-4.1-mini"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	required := map[string]string{
		"DISCORD_BOT_TOKEN":               c.DiscordBotToken,
		"DISCORD_SOURCE_CHANNEL_ID":       c.DiscordSourceChannelID,
		"DISCORD_NOTIFICATION_CHANNEL_ID": c.DiscordNotificationChanID,
		"NOTION_TOKEN":                    c.NotionToken,
		"NOTION_PARENT_PAGE_ID":           c.NotionParentPageID,
		"OPENAI_API_KEY":                  c.OpenAIAPIKey,
	}

	var missing []string
	for key, value := range required {
		if value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

var ErrNotImplemented = errors.New("not implemented")

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
