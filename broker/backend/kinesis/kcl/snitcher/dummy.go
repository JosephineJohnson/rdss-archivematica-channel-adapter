package snitcher

import (
	"sync"
)

// DummySnitcher is an implementation of Snitcher.
type DummySnitcher struct {
	state map[string]int
	mu    sync.RWMutex
}

// NewDummySnitcher returns a new DummySnitcher.
func NewDummySnitcher() *DummySnitcher {
	return &DummySnitcher{
		state: make(map[string]int),
	}
}

// CheckOwnership implements Snitcher.
func (s *DummySnitcher) CheckOwnership(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.state[key]
	return ok
}

// RegisterKey implements Snitcher.
func (s *DummySnitcher) RegisterKey(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[key] = 0
}
