# codex-sample-project

Weekly Discord digest system implemented in Go.

## What it does

- reads the previous 7 days of messages from a Discord channel
- creates a weekly Notion page
- generates an AI summary with OpenAI
- posts the Notion link and summary to a Discord notification channel

## Schedule

- every Saturday at 09:00 JST via GitHub Actions

## Required secrets

- `DISCORD_BOT_TOKEN`
- `DISCORD_SOURCE_CHANNEL_ID`
- `DISCORD_NOTIFICATION_CHANNEL_ID`
- `DISCORD_MENTION_USER_ID`
- `NOTION_TOKEN`
- `NOTION_PARENT_PAGE_ID`
- `OPENAI_API_KEY`
- `OPENAI_MODEL` optional, defaults to `gpt-4.1-mini`

## Local run

If you use `direnv`, place your secrets in `.env`, keep `.envrc` as-is, and run:

```bash
direnv allow
```

Then run the app normally:

```bash
go run ./cmd/weekly-digest
```

Use `.env.example` as the reference for required environment variables.

## Setup notes

- `DISCORD_MENTION_USER_ID` must be a numeric Discord user ID
- the Discord bot needs `VIEW_CHANNEL`, `READ_MESSAGE_HISTORY`, and `SEND_MESSAGES`
- the Notion parent page must be shared with the integration
- GitHub Actions cron `0 0 * * 6` corresponds to Saturday 09:00 JST

## Implementation scope

- source channel: `1007659965705625650`
- notification channel: `1482323925907144795`
- cadence: weekly
- output: Notion page plus Discord reminder

## Current limitations

- end-to-end execution has not been run yet because live credentials are not configured in this repository
- Discord mentions require the numeric user ID for `@tossy_yukky`
- the current Notion output stores messages as page blocks, so very large weekly volumes may eventually need pagination or truncation rules
