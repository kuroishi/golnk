package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchTitle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Example &amp; Test Title</title></head><body>hi</body></html>`))
	}))
	defer ts.Close()

	f := NewHTTPFetcher()
	got := f.FetchTitle(ts.URL)
	if got != "Example & Test Title" {
		t.Errorf("title = %q, want %q", got, "Example & Test Title")
	}
}

func TestFetchTitle_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	f := NewHTTPFetcher()
	if got := f.FetchTitle(ts.URL); got != "" {
		t.Errorf("title = %q, want empty", got)
	}
}

func TestFetchTitle_NoTitleTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>no title here</body></html>"))
	}))
	defer ts.Close()

	f := NewHTTPFetcher()
	if got := f.FetchTitle(ts.URL); got != "" {
		t.Errorf("title = %q, want empty", got)
	}
}

func TestFetchTitle_RequestError(t *testing.T) {
	f := NewHTTPFetcher()
	if got := f.FetchTitle("http://127.0.0.1:1/"); got != "" {
		t.Errorf("title = %q, want empty", got)
	}
}