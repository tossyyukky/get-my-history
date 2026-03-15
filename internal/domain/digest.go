package domain

import "time"

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
	Window    DigestWindow
	Messages  []Message
	Summary   string
	NotionURL string
}

func (w WeeklyDigest) MessageCount() int {
	return len(w.Messages)
}
