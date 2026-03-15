package app

import (
	"context"
	"fmt"
	"time"

	"github.com/tossy-yukky/codex-sample-project/internal/config"
	"github.com/tossy-yukky/codex-sample-project/internal/discord"
	"github.com/tossy-yukky/codex-sample-project/internal/domain"
	"github.com/tossy-yukky/codex-sample-project/internal/notion"
	"github.com/tossy-yukky/codex-sample-project/internal/openai"
)

type DiscordReader interface {
	FetchMessages(ctx context.Context, channelID string, window domain.DigestWindow) ([]domain.Message, error)
}

type NotionPublisher interface {
	CreateWeeklyPage(ctx context.Context, digest domain.WeeklyDigest) (string, error)
}

type Summarizer interface {
	Summarize(ctx context.Context, digest domain.WeeklyDigest) (string, error)
}

type DiscordNotifier interface {
	SendDigest(ctx context.Context, channelID, mentionUserID, notionURL, summary string) error
}

func Run(ctx context.Context, cfg config.Config) error {
	window := CurrentWeeklyWindow(time.Now())
	discordClient := discord.NewClient(cfg)
	notionClient := notion.NewClient(cfg)
	openAIClient := openai.NewClient(cfg)

	messages, err := discordClient.FetchMessages(ctx, cfg.DiscordSourceChannelID, window)
	if err != nil {
		return fmt.Errorf("fetch discord messages: %w", err)
	}

	digest := domain.WeeklyDigest{
		Window:   window,
		Messages: messages,
	}

	summary, err := openAIClient.Summarize(ctx, digest)
	if err != nil {
		return fmt.Errorf("summarize weekly digest: %w", err)
	}
	digest.Summary = summary

	notionURL, err := notionClient.CreateWeeklyPage(ctx, digest)
	if err != nil {
		return fmt.Errorf("create notion page: %w", err)
	}
	digest.NotionURL = notionURL

	if err := discordClient.SendDigest(
		ctx,
		cfg.DiscordNotificationChanID,
		cfg.DiscordMentionUserID,
		digest.NotionURL,
		digest.Summary,
	); err != nil {
		return fmt.Errorf("send discord digest: %w", err)
	}

	return nil
}

func CurrentWeeklyWindow(now time.Time) domain.DigestWindow {
	jst := time.FixedZone("JST", 9*60*60)
	localNow := now.In(jst)
	end := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 9, 0, 0, 0, jst)

	for end.Weekday() != time.Saturday || localNow.Before(end) {
		end = end.AddDate(0, 0, -1)
	}

	start := end.AddDate(0, 0, -7)
	return domain.DigestWindow{
		Start: start.UTC(),
		End:   end.UTC(),
	}
}
