package repo

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kuroishi/golnk/internal/domain"
)

func TestFileRepo_SaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	r, err := NewFileRepo(path)
	if err != nil {
		t.Fatalf("NewFileRepo: %v", err)
	}
	created := time.Now()
	if err := r.Save("abc123", &domain.Link{URL: "https://example.com/", Title: "Example", CreatedAt: created}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	r2, err := NewFileRepo(path)
	if err != nil {
		t.Fatalf("reload NewFileRepo: %v", err)
	}
	l, err := r2.Find("abc123")
	if err != nil {
		t.Fatalf("Find after reload: %v", err)
	}
	if l.URL != "https://example.com/" || l.Title != "Example" {
		t.Errorf("reloaded link = %+v", l)
	}
	if !l.CreatedAt.Equal(created) {
		t.Errorf("CreatedAt = %v, want %v", l.CreatedAt, created)
	}
}

func TestFileRepo_Conflict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	r, err := NewFileRepo(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Save("id", &domain.Link{URL: "https://a.example/"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Save("id", &domain.Link{URL: "https://b.example/"}); err != domain.ErrConflict {
		t.Errorf("err = %v, want domain.ErrConflict", err)
	}
}

func TestFileRepo_FindMissing(t *testing.T) {
	r, _ := NewFileRepo(filepath.Join(t.TempDir(), "data.json"))
	_, err := r.Find("nope")
	if err != domain.ErrNotFound {
		t.Errorf("err = %v, want domain.ErrNotFound", err)
	}
}

func TestFileRepo_EmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := NewFileRepo(path)
	if err != nil {
		t.Fatalf("NewFileRepo with empty file: %v", err)
	}
	if len(r.data) != 0 {
		t.Errorf("len(data) = %d, want 0", len(r.data))
	}
}