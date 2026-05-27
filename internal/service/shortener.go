package service

import (
	"errors"
	"net/url"
	"strings"

	"go.uber.org/fx"
)

type Shortener struct {
	storage Storage
}

func NewShortener(storage Storage) *Shortener {
	return &Shortener{storage: storage}
}

func (s *Shortener) Shorten(originalURL string) (string, error) {
	original := strings.TrimSpace(originalURL)

	if original == "" {
		return "", ErrInvalidURL
	}

	u, err := url.ParseRequestURI(original)
	if err != nil {
		return "", ErrInvalidURL
	}

	if u.Scheme == "" || u.Host == "" {
		return "", ErrInvalidURL
	}

	id, err := generateID()
	if err != nil {
		return "", err
	}

	if err := s.storage.Put(id, original); err != nil {
		return "", err
	}

	return id, nil
}

func (s *Shortener) Resolve(id string) (string, error) {
	originalURL, err := s.storage.Get(id)
	if errors.Is(err, ErrStorageNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func Provide() fx.Option {
	return fx.Provide(NewShortener)
}
