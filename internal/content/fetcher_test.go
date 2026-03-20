package content

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/tossy-yukky/codex-sample-project/internal/domain"
)

func TestExtractURLs(t *testing.T) {
	input := "see https://example.com/a and https://example.com/a, also https://example.com/b."
	got := ExtractURLs(input)

	if len(got) != 2 {
		t.Fatalf("expected 2 urls, got %d: %#v", len(got), got)
	}
	if got[0] != "https://example.com/a" || got[1] != "https://example.com/b" {
		t.Fatalf("unexpected urls: %#v", got)
	}
}

func TestFetchReferencesHTML(t *testing.T) {
	fetcher := NewFetcherWithClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"text/html; charset=utf-8"},
				},
				Body: io.NopCloser(strings.NewReader(`<html><head><title>Example Page</title></head><body><main>Hello <b>world</b>. This is a test page.</main></body></html>`)),
			}, nil
		}),
	})
	refs := fetcher.FetchReferences(context.Background(), []domain.Message{
		{ID: "1", Content: "check https://example.com/article"},
	})

	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
	if refs[0].Title != "Example Page" {
		t.Fatalf("unexpected title: %q", refs[0].Title)
	}
	if refs[0].Excerpt == "" {
		t.Fatal("expected non-empty excerpt")
	}
	if refs[0].Failed() {
		t.Fatalf("expected success, got failure: %q", refs[0].Error)
	}
}

func TestFetchReferencesUsesOpenGraphMetadata(t *testing.T) {
	fetcher := NewFetcherWithClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"text/html; charset=utf-8"},
				},
				Body: io.NopCloser(strings.NewReader(`<html><head><meta property="og:title" content="X Post Title"><meta property="og:description" content="Posted via metadata excerpt."><title>Fallback Title</title></head><body><main>Fallback body text</main></body></html>`)),
			}, nil
		}),
	})

	refs := fetcher.FetchReferences(context.Background(), []domain.Message{
		{ID: "1", Content: "check https://x.com/example/status/1"},
	})

	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
	if refs[0].Title != "X Post Title" {
		t.Fatalf("unexpected title: %q", refs[0].Title)
	}
	if refs[0].Excerpt != "Posted via metadata excerpt." {
		t.Fatalf("unexpected excerpt: %q", refs[0].Excerpt)
	}
}

func TestFetchReferencesRecordsFailure(t *testing.T) {
	fetcher := NewFetcherWithClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Header: http.Header{
					"Content-Type": []string{"text/plain; charset=utf-8"},
				},
				Body: io.NopCloser(strings.NewReader("forbidden")),
			}, nil
		}),
	})

	refs := fetcher.FetchReferences(context.Background(), []domain.Message{
		{ID: "1", Content: "check https://example.com/private"},
	})

	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
	if !refs[0].Failed() {
		t.Fatal("expected failed reference")
	}
	if refs[0].Error == "" {
		t.Fatal("expected error message")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
