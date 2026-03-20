# Weekly Discord Digest Plan

## Goal

Build a private GitHub Actions based system in Go that:

- reads messages posted in Discord channel `1007659965705625650`
- aggregates the previous 7 days of posts at each execution time
- creates one weekly Notion page
- generates an AI summary using OpenAI API
- posts the Notion link and summary to Discord channel `1482323925907144795`
- mentions the target user in the reminder message

## Constraints

- Repository visibility: private
- Execution platform: GitHub Actions
- Language: Go
- Raw message storage in GitHub: not required
- Source of truth for raw content: Discord

## Decisions

- Schedule: every Saturday at 09:00 JST
- Notion format: one page per week
- Summary provider: OpenAI API
- Discord read access: Bot token required
- Discord write access: Bot token preferred for one credential path
- Mention target: user mention, final Discord user ID still needs to be configured

## Phases

### Phase 1: Project scaffold

- initialize Go module
- define package layout and configuration model
- add README with setup and architecture notes
- create progress tracking files

### Phase 2: Discord ingestion

- implement Discord API client for reading channel history
- filter messages to the last 7 days relative to the execution time
- normalize messages into internal digest items

### Phase 3: Notion publishing

- implement Notion page creation for a weekly digest
- render weekly content in a readable format
- return created page URL for downstream notification

### Phase 4: AI summarization

- implement OpenAI based summarization over collected Discord posts
- constrain output for Discord readability

### Phase 5: Reminder delivery

- implement Discord notification message to channel `1482323925907144795`
- support optional mention via configured user ID
- include Notion URL and summary in final post

### Phase 6: Automation

- add GitHub Actions workflow with Saturday 09:00 JST schedule
- document required GitHub Secrets and operational setup

### Phase 7: Verification

- add unit tests for date windowing and formatting
- run local checks where possible
- document known limitations and follow-up tasks

## Open Items

- Discord mention user ID for `@tossy_yukky`
- Notion parent page or database target ID
- Discord bot installation and permissions in the server
