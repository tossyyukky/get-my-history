package domain

import "testing"

func TestMessageHasMissingContent(t *testing.T) {
	msg := Message{Content: "   "}
	if !msg.HasMissingContent() {
		t.Fatal("expected missing content to be detected")
	}
}

func TestWeeklyDigestMessagesWithMissingContent(t *testing.T) {
	digest := WeeklyDigest{
		Messages: []Message{
			{ID: "1", Content: ""},
			{ID: "2", Content: "hello"},
		},
	}

	got := digest.MessagesWithMissingContent()
	if len(got) != 1 {
		t.Fatalf("expected 1 missing-content message, got %d", len(got))
	}
	if got[0].ID != "1" {
		t.Fatalf("unexpected message id: %s", got[0].ID)
	}
}
