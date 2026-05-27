package service

type Storage interface {
	Get(id string) (string, error)
	Put(id, originalURL string) error
}
