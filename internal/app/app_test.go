package app

import (
	"testing"
	"time"
)

func TestCurrentWeeklyWindowAfterScheduledTime(t *testing.T) {
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	window := CurrentWeeklyWindow(now)

	assertTimeEqual(t, "start", window.Start, time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC))
	assertTimeEqual(t, "end", window.End, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC))
}

func TestCurrentWeeklyWindowBeforeScheduledTime(t *testing.T) {
	now := time.Date(2026, 3, 21, 8, 30, 0, 0, time.FixedZone("JST", 9*60*60))
	window := CurrentWeeklyWindow(now)

	assertTimeEqual(t, "start", window.Start, time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC))
	assertTimeEqual(t, "end", window.End, time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC))
}

func assertTimeEqual(t *testing.T, label string, got, want time.Time) {
	t.Helper()
	if !got.Equal(want) {
		t.Fatalf("%s mismatch: got=%s want=%s", label, got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}
