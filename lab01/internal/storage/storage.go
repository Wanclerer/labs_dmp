// Package storage хранит соответствие короткого кода и исходной ссылки.
package storage

import (
	"errors"
	"sync"
)

// ErrCodeTaken возвращается, если код уже занят другой ссылкой.
var ErrCodeTaken = errors.New("code already taken")

// Storage — хранилище ссылок.
type Storage interface {
	Save(code, url string) error
	Lookup(code string) (url string, ok bool)
}

type memoryStorage struct {
	mu    sync.RWMutex
	items map[string]string
}

// New создаёт хранилище в памяти.
func New() Storage {
	return &memoryStorage{
		items: make(map[string]string),
	}
}

// Save сохраняет ссылку под кодом.
// Проверка доступности кода и запись выполняются неделимо под Lock().
func (s *memoryStorage) Save(code, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[code]; exists {
		return ErrCodeTaken
	}

	s.items[code] = url
	return nil
}

// Lookup возвращает ссылку по коду.
func (s *memoryStorage) Lookup(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.items[code]
	return url, ok
}