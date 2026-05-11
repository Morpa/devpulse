// Cria todas as dependências, liga-as entre si e arranca o servidor Gin.
// Estrutura simples: config → store → services → handlers → router.
package main

import (
	"log"
	"strconv"
	"time"

	"github.com/Morpa/devpulse/api/internal/config"
	"github.com/Morpa/devpulse/api/internal/handlers"
	"github.com/Morpa/devpulse/api/internal/middleware"
	"github.com/Morpa/devpulse/api/internal/services"
	"github.com/Morpa/devpulse/api/internal/sse"
	"github.com/Morpa/devpulse/api/internal/store"
	"github.com/gin-gonic/gin"
)

func main() {
	// ── Configuração ──────────────────────────────────────────────────────────
	cfg := config.Load()

	// Gin em Release mode — sem logs de debug coloridos (mais limpo no Docker)
	gin.SetMode(gin.ReleaseMode)

	// ── Dependências ──────────────────────────────────────────────────────────
	repoStore := store.New()
	ghClient := services.NewGitHubClient(cfg.GitHubToken)
	sseBroker := sse.New()
	go sseBroker.Run() // event-loop do broker numa goroutine dedicada

	// ── Scheduler de refresh periódico ────────────────────────────────────────
	refreshMins, err := strconv.Atoi(cfg.RefreshMins)
	if err != nil || refreshMins < 1 {
		refreshMins = 5
	}
	sched := services.NewScheduler(repoStore, ghClient, sseBroker, time.Duration(refreshMins)*time.Minute)
	go sched.Start()

	// ── Router Gin ────────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Health check — usado pelo Docker/k8s para saber se o processo está vivo
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Handlers de repos
	repoHandler := handlers.NewRepoHandler(repoStore, ghClient, sseBroker)

	api := r.Group("/api")
	{
		api.GET("/repos", repoHandler.List)
		api.POST("/repos", repoHandler.Track)
		api.DELETE("/repos/:owner/:repo", repoHandler.Untrack)
		api.GET("/events", sseBroker.Handler()) // SSE stream
	}

	// ── Arrancar ──────────────────────────────────────────────────────────────
	log.Printf("🚀 DevPulse API a escutar em :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("erro ao arrancar: %v", err)
	}
}
