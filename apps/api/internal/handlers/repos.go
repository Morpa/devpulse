// Cada handler faz exactamente três coisas:
//  1. Lê o input (params, body)
//  2. Chama o serviço
//  3. Devolve a resposta JSON
//
// Sem lógica de negócio aqui — isso fica nos services.
package handlers

import (
	"net/http"
	"strings"

	"github.com/Morpa/devpulse/api/internal/models"
	"github.com/Morpa/devpulse/api/internal/services"
	"github.com/Morpa/devpulse/api/internal/sse"
	"github.com/Morpa/devpulse/api/internal/store"
	"github.com/gin-gonic/gin"
)

// RepoHandler agrupa os handlers e as dependências que precisam.
type RepoHandler struct {
	store  *store.Store
	github *services.GitHubClient
	broker *sse.Broker
}

// NewRepoHandler cria o handler com as dependências injectadas.
func NewRepoHandler(s *store.Store, gh *services.GitHubClient, broker *sse.Broker) *RepoHandler {
	return &RepoHandler{store: s, github: gh, broker: broker}
}

// List devolve todos os repos tracked ordenados por health score.
// GET /api/repos
func (h *RepoHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.All())
}

// Track adiciona um repo ao tracking e faz o primeiro fetch de métricas.
// POST /api/repos   body: {"owner": "golang", "repo": "go"}
func (h *RepoHandler) Track(c *gin.Context) {
	var body struct {
		Owner string `json:"owner" binding:"required"`
		Repo  string `json:"repo"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner e repo são obrigatórios"})
		return
	}

	owner := strings.TrimSpace(body.Owner)
	name := strings.TrimSpace(body.Repo)
	key := owner + "/" + name

	if h.store.Exists(key) {
		c.JSON(http.StatusConflict, gin.H{"error": "repositório já está a ser monitorizado"})
		return
	}

	// Fetch das métricas — pode demorar até ~1s (5 pedidos paralelos ao GitHub)
	snap, err := h.github.FetchRepo(c.Request.Context(), owner, name)
	if err != nil {
		status := http.StatusBadGateway
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "rate limit") {
			status = http.StatusTooManyRequests
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	h.store.Set(key, snap)
	h.broker.Publish(models.SSEEvent{Type: "repo_added", Payload: snap})

	c.JSON(http.StatusCreated, snap)
}

// Untrack remove um repo do tracking.
// DELETE /api/repos/:owner/:repo
func (h *RepoHandler) Untrack(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	key := owner + "/" + repo

	if !h.store.Exists(key) {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	h.store.Delete(key)
	h.broker.Publish(models.SSEEvent{Type: "repo_removed", Payload: gin.H{"key": key}})

	c.Status(http.StatusNoContent)
}
