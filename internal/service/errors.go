package service

import "errors"

var ErrNotFound = errors.New("service: id not found")
var ErrInvalidURL = errors.New("service: invalid url")

// ErrStorageNotFound — sentinel-контракт интерфейса Storage
var ErrStorageNotFound = errors.New("storage: id not found")
