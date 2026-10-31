package fetcher

import (
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var titleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

type HTTPFetcher struct {
	client *http.Client
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

func (f *HTTPFetcher) FetchTitle(rawURL string) string {
	resp, err := f.client.Get(rawURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return ""
	}
	m := titleRe.FindSubmatch(b)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(html.UnescapeString(string(m[1])))
}