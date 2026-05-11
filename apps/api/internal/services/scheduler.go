// Corre num ticker e re-fetcha todos os repos tracked usando um worker pool
// para não ultrapassar os rate limits do GitHub (máx. 3 em simultâneo).
package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Morpa/devpulse/api/internal/models"
	"github.com/Morpa/devpulse/api/internal/sse"
	"github.com/Morpa/devpulse/api/internal/store"
)

const maxWorkers = 3 // máximo de pedidos paralelos ao GitHub

// Scheduler re-fetcha todos os repos num intervalo fixo.
type Scheduler struct {
	store    *store.Store
	github   *GitHubClient
	broker   *sse.Broker
	interval time.Duration
}

// NewScheduler cria o scheduler. Chama Start() numa goroutine para o activar.
func NewScheduler(s *store.Store, gh *GitHubClient, broker *sse.Broker, interval time.Duration) *Scheduler {
	return &Scheduler{
		store:    s,
		github:   gh,
		broker:   broker,
		interval: interval,
	}
}

// Start bloqueia para sempre, refrescando a cada s.interval.
// Deve correr numa goroutine dedicada.
func (s *Scheduler) Start() {
	log.Printf("📅 Scheduler activo — intervalo: %s", s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for range ticker.C {
		s.RefreshAll()
	}
}

// RefreshAll re-fetcha todos os repos com um worker pool de tamanho fixo.
// Erros individuais marcam o repo como "stale" mas não param os restantes.
func (s *Scheduler) RefreshAll() {
	keys := s.store.Keys()
	if len(keys) == 0 {
		return
	}

	log.Printf("🔄 Refrescando %d repos…", len(keys))

	// Semáforo via canal buffered — só maxWorkers goroutines correm ao mesmo tempo
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, key := range keys {
		wg.Add(1)
		sem <- struct{}{} // adquire slot

		go func(k string) {
			defer wg.Done()
			defer func() { <-sem }() // liberta slot

			s.refreshOne(k)
		}(key)
	}

	wg.Wait()
	log.Printf("✅ Refresh completo")
}

// refreshOne re-fetcha um único repo e actualiza o store + SSE.
func (s *Scheduler) refreshOne(key string) {
	snap, exists := s.store.Get(key)
	if !exists {
		return // foi removido entre Keys() e agora
	}

	updated, err := s.github.FetchRepo(context.Background(), snap.Owner, snap.Name)
	if err != nil {
		log.Printf("❌ refresh %s: %v", key, err)
		// Manter dados antigos mas marcar como stale
		snap.Status = "stale"
		s.store.Set(key, snap)
		s.broker.Publish(models.SSEEvent{Type: "repo_updated", Payload: snap})
		return
	}

	s.store.Set(key, updated)
	s.broker.Publish(models.SSEEvent{Type: "repo_updated", Payload: updated})
	log.Printf("✔ %s — health: %d", key, updated.HealthScore)
}
