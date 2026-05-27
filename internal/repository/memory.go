package repository

import (
	"github.com/SZabrodskii/url-shortener/internal/service"
	syncmap "github.com/chloyka/sync-map-generic"
	"go.uber.org/fx"
)

type MemoryStorage struct {
	data syncmap.KVMap[string, string]
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (s *MemoryStorage) Get(id string) (string, error) {
	value, ok := s.data.Load(id)
	if !ok {
		return "", service.ErrStorageNotFound
	}

	return *value, nil
}

func (s *MemoryStorage) Put(id, originalURL string) error {
	s.data.Store(id, &originalURL)
	return nil
}

func Provide() fx.Option {
	return fx.Provide(
		fx.Annotate(
			NewMemoryStorage,
			fx.As(new(service.Storage)),
		),
	)
}
