// Padrão event-loop: uma goroutine é dona do mapa de clientes.
// Todas as outras interagem apenas via canais — sem mutex no hot path.
//
//	Publish(event) ──► canal buffered ──► event-loop ──► fan-out ──► clientChan ──► browser
package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Morpa/devpulse/api/internal/models"
	"github.com/gin-gonic/gin"
)

// client é o canal dedicado a cada tab de browser ligada.
type client chan string

// Broker gere as ligações SSE e distribui eventos.
type Broker struct {
	register   chan client
	unregister chan client
	publish    chan models.SSEEvent
	clients    map[client]struct{}
}

// New cria o Broker. Chama Run() numa goroutine para o activar.
func New() *Broker {
	return &Broker{
		register:   make(chan client),
		unregister: make(chan client),
		publish:    make(chan models.SSEEvent, 64), // buffered — Publish() não bloqueia
		clients:    make(map[client]struct{}),
	}
}

// Run é o event-loop — corre numa goroutine dedicada.
// É a única goroutine que lê/escreve no mapa clients.
func (b *Broker) Run() {
	for {
		select {
		case c := <-b.register:
			b.clients[c] = struct{}{}
			log.Printf("SSE: +1 cliente (total: %d)", len(b.clients))

		case c := <-b.unregister:
			delete(b.clients, c)
			close(c)
			log.Printf("SSE: -1 cliente (total: %d)", len(b.clients))

		case event := <-b.publish:
			// Serializa uma vez e manda para todos
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("SSE: erro a serializar: %v", err)
				continue
			}
			msg := string(data)
			for c := range b.clients {
				select {
				case c <- msg:
				default:
					// Canal cheio — cliente lento, descarta para não bloquear os outros
				}
			}
		}
	}
}

// Publish envia um evento para todos os browsers ligados. Thread-safe.
func (b *Broker) Publish(event models.SSEEvent) {
	b.publish <- event
}

// Handler devolve o gin.HandlerFunc para GET /api/events.
// Mantém a ligação aberta e faz stream dos eventos até o browser fechar.
func (b *Broker) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
			return
		}

		// Headers SSE obrigatórios
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no") // desactiva buffer do nginx/caddy

		ch := make(client, 32)
		b.register <- ch
		defer func() { b.unregister <- ch }()

		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				// Formato SSE: "data: <json>\n\n"
				fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
				flusher.Flush()

			case <-c.Request.Context().Done():
				// Browser fechou o tab
				return
			}
		}
	}
}
