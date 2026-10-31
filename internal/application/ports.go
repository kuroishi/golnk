package application

import "github.com/kuroishi/golnk/internal/domain"

type LinkRepository interface {
	Save(id string, link *domain.Link) error
	Find(id string) (*domain.Link, error)
}

type TitleFetcher interface {
	FetchTitle(rawURL string) string
}

type IDGenerator interface {
	Generate() string
}