package content

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tossy-yukky/codex-sample-project/internal/domain"
)

var urlPattern = regexp.MustCompile(`https?://[^\s<>"']+`)

type Fetcher struct {
	httpClient *http.Client
}

func NewFetcher() *Fetcher {
	return NewFetcherWithClient(&http.Client{
		Timeout: 15 * time.Second,
	})
}

func NewFetcherWithClient(httpClient *http.Client) *Fetcher {
	return &Fetcher{
		httpClient: httpClient,
	}
}

func (f *Fetcher) FetchReferences(ctx context.Context, messages []domain.Message) []domain.ReferenceContent {
	seen := map[string]struct{}{}
	var refs []domain.ReferenceContent

	for _, msg := range messages {
		urls := ExtractURLs(msg.Content)
		for _, rawURL := range urls {
			if _, ok := seen[rawURL]; ok {
				continue
			}
			seen[rawURL] = struct{}{}

			ref, err := f.fetchReference(ctx, msg.ID, rawURL)
			if err != nil {
				continue
			}
			refs = append(refs, ref)
		}
	}

	slices.SortFunc(refs, func(a, b domain.ReferenceContent) int {
		return strings.Compare(a.URL, b.URL)
	})

	return refs
}

func ExtractURLs(text string) []string {
	matches := urlPattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := map[string]struct{}{}
	urls := make([]string, 0, len(matches))
	for _, match := range matches {
		clean := strings.TrimRight(match, ".,);!?]")
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		urls = append(urls, clean)
	}

	return urls
}

func (f *Fetcher) fetchReference(ctx context.Context, messageID, rawURL string) (domain.ReferenceContent, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return domain.ReferenceContent{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "get-my-history/1.0")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return domain.ReferenceContent{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return domain.ReferenceContent{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "text/plain") {
		return domain.ReferenceContent{}, fmt.Errorf("unsupported content type: %s", contentType)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return domain.ReferenceContent{}, fmt.Errorf("read body: %w", err)
	}

	title, excerpt := summarizeDocument(body)
	if excerpt == "" {
		return domain.ReferenceContent{}, fmt.Errorf("empty excerpt")
	}

	return domain.ReferenceContent{
		SourceMessageID: messageID,
		URL:             rawURL,
		Title:           title,
		Excerpt:         excerpt,
	}, nil
}

func summarizeDocument(body []byte) (string, string) {
	text := string(body)
	title := extractTitle(text)
	text = stripTagBlock(text, "script")
	text = stripTagBlock(text, "style")
	text = stripTags(text)
	text = html.UnescapeString(text)
	text = normalizeWhitespace(text)

	if title == "" {
		title = firstLine(text)
	}

	return truncate(title, 160), truncate(text, 1200)
}

func extractTitle(body string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	m := re.FindStringSubmatch(body)
	if len(m) < 2 {
		return ""
	}
	return normalizeWhitespace(html.UnescapeString(stripTags(m[1])))
}

func stripTagBlock(body, tag string) string {
	re := regexp.MustCompile(`(?is)<` + tag + `[^>]*>.*?</` + tag + `>`)
	return re.ReplaceAllString(body, " ")
}

func stripTags(body string) string {
	re := regexp.MustCompile(`(?s)<[^>]+>`)
	return re.ReplaceAllString(body, " ")
}

func normalizeWhitespace(value string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(value), " "))
}

func firstLine(value string) string {
	if value == "" {
		return ""
	}
	return strings.SplitN(value, "\n", 2)[0]
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	var b bytes.Buffer
	b.WriteString(value[:limit-3])
	b.WriteString("...")
	return b.String()
}
