package domain

import (
	"strings"
	"time"
)

type Message struct {
	ID        string
	AuthorID  string
	Author    string
	Content   string
	Timestamp time.Time
	URL       string
}

type DigestWindow struct {
	Start time.Time
	End   time.Time
}

type WeeklyDigest struct {
	Window     DigestWindow
	Messages   []Message
	References []ReferenceContent
	Summary    string
	NotionURL  string
}

func (w WeeklyDigest) MessageCount() int {
	return len(w.Messages)
}

func (w WeeklyDigest) MessagesWithMissingContent() []Message {
	result := make([]Message, 0, len(w.Messages))
	for _, msg := range w.Messages {
		if msg.HasMissingContent() {
			result = append(result, msg)
		}
	}
	return result
}

type ReferenceContent struct {
	SourceMessageID string
	URL             string
	Title           string
	Excerpt         string
	Error           string
}

func (r ReferenceContent) Failed() bool {
	return r.Error != ""
}

func (m Message) HasMissingContent() bool {
	return strings.TrimSpace(m.Content) == ""
}
