package repo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/kuroishi/golnk/internal/domain"
)

type FileRepo struct {
	mu       sync.Mutex
	data     map[string]*domain.Link
	dataFile string
}

func NewFileRepo(path string) (*FileRepo, error) {
	r := &FileRepo{data: map[string]*domain.Link{}, dataFile: path}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *FileRepo) Save(id string, link *domain.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; ok {
		return domain.ErrConflict
	}
	cp := *link
	r.data[id] = &cp
	return r.saveLocked()
}

func (r *FileRepo) Find(id string) (*domain.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return l, nil
}

func (r *FileRepo) load() error {
	b, err := os.ReadFile(r.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(b) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return json.Unmarshal(b, &r.data)
}

func (r *FileRepo) saveLocked() error {
	b, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(r.dataFile)
	tmp, err := os.CreateTemp(dir, ".data-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, r.dataFile)
}