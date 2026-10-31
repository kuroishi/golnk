package application

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kuroishi/golnk/internal/domain"
)

type fakeRepo struct {
	mu   sync.Mutex
	data map[string]*domain.Link
}

func newFakeRepo() *fakeRepo {
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

type fakeFetcher struct {
	title string
}

func (f fakeFetcher) FetchTitle(string) string { return f.title }

type fakeGen struct {
	mu   sync.Mutex
	ids  []string
	next int
}

func (g *fakeGen) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.ids != nil {
		id := g.ids[g.next%len(g.ids)]
		g.next++
		return id
	}
	id := fmt.Sprintf("id%03d", g.next)
	g.next++
	return id
}

func TestShorten_Valid(t *testing.T) {
	repo := newFakeRepo()
	s := NewShortener(repo, fakeFetcher{title: "Example"}, &fakeGen{})

	id, link, err := s.Shorten("https://example.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Error("id is empty")
	}
	if link == nil {
		t.Fatal("link is nil")
	}
	if link.URL != "https://example.com/" {
		t.Errorf("URL = %q", link.URL)
	}
	if link.Title != "Example" {
		t.Errorf("Title = %q, want Example", link.Title)
	}
	if link.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if _, err := repo.Find(id); err != nil {
		t.Errorf("saved link not found: %v", err)
	}
}

func TestShorten_InvalidURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"ftp", "ftp://example.com/"},
		{"no host", "https://"},
		{"garbage", "://bad"},
		{"empty", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewShortener(newFakeRepo(), fakeFetcher{}, &fakeGen{})
			_, _, err := s.Shorten(tc.url)
			if !errors.Is(err, domain.ErrInvalidURL) {
				t.Errorf("err = %v, want ErrInvalidURL", err)
			}
		})
	}
}

func TestShorten_CollisionRetries(t *testing.T) {
	repo := newFakeRepo()
	if err := repo.Save("id000", &domain.Link{URL: "https://old.example/"}); err != nil {
		t.Fatal(err)
	}
	s := NewShortener(repo, fakeFetcher{title: "T"}, &fakeGen{ids: []string{"id000", "id001"}})

	id, _, err := s.Shorten("https://example.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "id001" {
		t.Errorf("id = %q, want id001", id)
	}
}

func TestShorten_TitleFetchFailed(t *testing.T) {
	s := NewShortener(newFakeRepo(), fakeFetcher{title: ""}, &fakeGen{})

	_, link, err := s.Shorten("https://example.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.Title != "" {
		t.Errorf("Title = %q, want empty", link.Title)
	}
}

func TestResolve(t *testing.T) {
	repo := newFakeRepo()
	if err := repo.Save("abc123", &domain.Link{URL: "https://example.com/", Title: "Example", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	s := NewShortener(repo, fakeFetcher{}, &fakeGen{})

	l, err := s.Resolve("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.URL != "https://example.com/" {
		t.Errorf("URL = %q", l.URL)
	}

	_, err = s.Resolve("nope")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestShorten_Concurrent(t *testing.T) {
	repo := newFakeRepo()
	s := NewShortener(repo, fakeFetcher{title: "T"}, &fakeGen{})

	var wg sync.WaitGroup
	const n = 20
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			url := fmt.Sprintf("https://example.com/%d", i)
			if _, _, err := s.Shorten(url); err != nil {
				t.Errorf("Shorten(%s): %v", url, err)
			}
		}(i)
	}
	wg.Wait()
	if len(repo.data) != n {
		t.Errorf("len(repo.data) = %d, want %d", len(repo.data), n)
	}
}