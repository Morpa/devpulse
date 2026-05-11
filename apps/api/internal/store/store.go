// Usa sync.RWMutex: múltiplos leitores simultâneos, escritas exclusivas.
// Para persistência real bastaria trocar por SQLite sem mudar os callers.

package store

import (
	"sort"
	"sync"

	"github.com/Morpa/devpulse/api/internal/models"
)

// Store guarda todos os repos tracked em memória.
type Store struct {
	mu    sync.RWMutex
	repos map[string]*models.RepoSnapshot
}

// New cria um Store vazio.
func New() *Store {
	return &Store{
		repos: make(map[string]*models.RepoSnapshot),
	}
}

// Set guarda ou actualiza um snapshot.
func (s *Store) Set(key string, snap *models.RepoSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repos[key] = snap
}

// Get devolve um snapshot pelo key. Segundo valor indica se existe.
func (s *Store) Get(key string) (*models.RepoSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.repos[key]
	return snap, ok
}

// Exists verifica se um key está tracked.
func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.repos[key]
	return ok
}

// Delete remove um repo.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.repos, key)
}

// All devolve todos os repos ordenados por health score descendente.
func (s *Store) All() []*models.RepoSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*models.RepoSnapshot, 0, len(s.repos))
	for _, snap := range s.repos {
		list = append(list, snap)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].HealthScore > list[j].HealthScore
	})

	return list
}

// Keys devolve todos os keys tracked — usado pelo scheduler.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.repos))
	for k := range s.repos {
		keys = append(keys, k)
	}
	return keys
}
