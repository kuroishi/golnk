package application

import (
	"errors"

	"github.com/kuroishi/golnk/internal/domain"
)

type Shortener struct {
	repo    LinkRepository
	fetcher TitleFetcher
	gen     IDGenerator
}

func NewShortener(repo LinkRepository, fetcher TitleFetcher, gen IDGenerator) *Shortener {
	return &Shortener{repo: repo, fetcher: fetcher, gen: gen}
}

func (s *Shortener) Shorten(rawURL string) (id string, link *domain.Link, err error) {
	link, err = domain.NewLink(rawURL)
	if err != nil {
		return "", nil, err
	}
	link.SetTitle(s.fetcher.FetchTitle(rawURL))
	for {
		id = s.gen.Generate()
		if err := s.repo.Save(id, link); err == nil {
			return id, link, nil
		} else if !errors.Is(err, domain.ErrConflict) {
			return "", nil, err
		}
	}
}

func (s *Shortener) Resolve(id string) (*domain.Link, error) {
	return s.repo.Find(id)
}