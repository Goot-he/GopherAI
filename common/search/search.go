package search

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	blockRe   = regexp.MustCompile(`<li class="b_algo"[^>]*>([\s\S]*?)</li>`)
	titleRe   = regexp.MustCompile(`<h2[^>]*><a[^>]*href="([^"]+)"[^>]*>([\s\S]*?)</a></h2>`)
	snippetRe = regexp.MustCompile(`<p[^>]*class="[^"]*b_lineclamp[^"]*"[^>]*>([\s\S]*?)</p>`)
	tagRe     = regexp.MustCompile(`<[^>]*>`)
	spaceRe   = regexp.MustCompile(`\s+`)
)

type SearchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
}

func Search(query string, maxResults int) ([]SearchResult, error) {
	if maxResults <= 0 {
		maxResults = 5
	}

	bingURL := fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(query))
	req, err := http.NewRequest("GET", bingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	results := parseBing(string(body), maxResults)
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for: %s", query)
	}

	return results, nil
}

func parseBing(htmlStr string, maxResults int) []SearchResult {
	var results []SearchResult
	blocks := blockRe.FindAllStringSubmatch(htmlStr, -1)

	for i := 0; i < len(blocks) && len(results) < maxResults; i++ {
		content := blocks[i][1]

		tm := titleRe.FindStringSubmatch(content)
		if tm == nil || len(tm) < 3 {
			continue
		}

		r := SearchResult{
			URL:   cleanURL(tm[1]),
			Title: cleanText(tm[2]),
		}

		sm := snippetRe.FindStringSubmatch(content)
		if len(sm) > 1 {
			r.Snippet = cleanText(sm[1])
		}

		if r.Title != "" && r.URL != "" {
			results = append(results, r)
		}
	}

	return results
}

func cleanURL(raw string) string {
	return strings.ReplaceAll(strings.TrimSpace(raw), "&amp;", "&")
}

func cleanText(raw string) string {
	r := strings.TrimSpace(raw)
	r = html.UnescapeString(r)
	r = tagRe.ReplaceAllString(r, "")
	r = spaceRe.ReplaceAllString(r, " ")
	return strings.TrimSpace(r)
}

func FormatResults(results []SearchResult) string {
	if len(results) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.Title))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", r.Snippet))
		}
		sb.WriteString(fmt.Sprintf("    URL: %s\n\n", r.URL))
	}
	return sb.String()
}
