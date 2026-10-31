package domain

import (
	"net/url"
	"time"
)

type Link struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

func NewLink(rawURL string) (*Link, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, ErrInvalidURL
	}
	return &Link{URL: rawURL, CreatedAt: time.Now()}, nil
}

func (l *Link) SetTitle(title string) {
	l.Title = title
}