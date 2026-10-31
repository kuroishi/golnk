package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/kuroishi/golnk/internal/application"
	"github.com/kuroishi/golnk/internal/domain"
)

type fakeRepo struct {
	mu   sync.Mutex
	data map[string]*domain.Link
}

func newFake() *fakeRepo {
	return &fakeRepo{data: map[string]*domain.Link{}}
}

func (f *fakeRepo) Save(id string, link *domain.Link) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.data[id]; ok {
		return domain.ErrConflict
	}
	cp := *link
	f.data[id] = &cp
	return nil
}

func (f *fakeRepo) Find(id string) (*domain.Link, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return l, nil
}

type fakeFetcher struct{}

func (fakeFetcher) FetchTitle(string) string { return "Example Title" }

type fakeGen struct{ n int }

func (g *fakeGen) Generate() string {
	g.n++
	return fmt.Sprintf("id%03d", g.n)
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	s := application.NewShortener(newFake(), fakeFetcher{}, &fakeGen{})
	h := NewHandler(s)
	mux := http.NewServeMux()
	h.Register(mux)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func TestShorten_OK(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.PostForm(ts.URL+"/api/shorten", url.Values{"url": {"https://example.com/"}})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		ShortURL string `json:"short_url"`
		Title    string `json:"title"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.ShortURL == "" {
		t.Error("short_url is empty")
	}
	if body.Title != "Example Title" {
		t.Errorf("title = %q", body.Title)
	}
}

func TestShorten_InvalidURL(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.PostForm(ts.URL+"/api/shorten", url.Values{"url": {"ftp://x.example/"}})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestShorten_MethodNotAllowed(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/shorten")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}

func TestRedirect_OK(t *testing.T) {
	ts := newTestServer(t)

	if _, err := http.PostForm(ts.URL+"/api/shorten", url.Values{"url": {"https://example.com/page"}}); err != nil {
		t.Fatal(err)
	}

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(ts.URL + "/id001")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("status = %d, want 302", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "https://example.com/page" {
		t.Errorf("Location = %q", loc)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/nope")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRoot_OK(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestShorten_RepoError(t *testing.T) {
	s := application.NewShortener(&errRepo{newFake()}, fakeFetcher{}, &fakeGen{})
	h := NewHandler(s)
	mux := http.NewServeMux()
	h.Register(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.PostForm(ts.URL+"/api/shorten", url.Values{"url": {"https://example.com/"}})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

type errRepo struct {
	*fakeRepo
}

func (e *errRepo) Save(id string, link *domain.Link) error {
	return errors.New("disk full")
}