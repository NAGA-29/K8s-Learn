package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Message represents a WebSocket message.
type Message struct {
	Type      string `json:"type"`
	Data      string `json:"data"`
	From      string `json:"from"`
	Timestamp string `json:"timestamp"`
}

// Hub manages all active WebSocket connections and message broadcasting.
type Hub struct {
	mu           sync.RWMutex
	clients      map[string]*Client
	messageCount int64
	rdb          *redis.Client
	redisEnabled bool
}

// Client represents a single WebSocket connection.
type Client struct {
	id   string
	conn *websocket.Conn
	hub  *Hub
	send chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func newHub() *Hub {
	h := &Hub{
		clients: make(map[string]*Client),
	}

	// Set up Redis if REDIS_URL is provided
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("Failed to parse REDIS_URL: %v, falling back to in-memory broadcast", err)
		} else {
			h.rdb = redis.NewClient(opts)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := h.rdb.Ping(ctx).Err(); err != nil {
				log.Printf("Failed to connect to Redis: %v, falling back to in-memory broadcast", err)
				h.rdb = nil
			} else {
				h.redisEnabled = true
				log.Println("Redis Pub/Sub enabled")
				go h.subscribeRedis()
			}
		}
	} else {
		log.Println("REDIS_URL not set, using in-memory broadcast")
	}

	return h
}

const redisChannel = "ws:broadcast"

// subscribeRedis listens for messages from Redis Pub/Sub and delivers them to local clients.
func (h *Hub) subscribeRedis() {
	ctx := context.Background()
	sub := h.rdb.Subscribe(ctx, redisChannel)
	defer sub.Close()

	ch := sub.Channel()
	for msg := range ch {
		h.broadcastLocal([]byte(msg.Payload))
	}
}

// broadcastLocal sends data to all locally connected clients.
func (h *Hub) broadcastLocal(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		select {
		case c.send <- data:
		default:
			// Client send buffer full, skip
		}
	}
}

// broadcast sends a message to all clients, using Redis if enabled.
func (h *Hub) broadcast(data []byte) {
	h.mu.Lock()
	h.messageCount++
	h.mu.Unlock()

	if h.redisEnabled {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.rdb.Publish(ctx, redisChannel, data).Err(); err != nil {
			log.Printf("Redis publish error: %v, falling back to local broadcast", err)
			h.broadcastLocal(data)
		}
	} else {
		h.broadcastLocal(data)
	}
}

func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.id] = c
	log.Printf("Client connected: %s (total: %d)", c.id, len(h.clients))
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c.id]; ok {
		close(c.send)
		delete(h.clients, c.id)
		log.Printf("Client disconnected: %s (total: %d)", c.id, len(h.clients))
	}
}

func (h *Hub) connectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) getMessageCount() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.messageCount
}

func generateClientID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// readPump reads messages from the WebSocket connection and broadcasts them.
func (c *Client) readPump() {
	defer func() {
		c.hub.removeClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("Read error from %s: %v", c.id, err)
			}
			break
		}

		// Try to parse incoming message
		var incoming Message
		if err := json.Unmarshal(raw, &incoming); err != nil {
			// If not valid JSON, treat raw text as the data field
			incoming = Message{
				Type: "message",
				Data: string(raw),
			}
		}

		// Fill in server-side fields
		incoming.From = c.id
		incoming.Timestamp = time.Now().UTC().Format(time.RFC3339)
		if incoming.Type == "" {
			incoming.Type = "message"
		}

		out, _ := json.Marshal(incoming)
		c.hub.broadcast(out)
	}
}

// writePump writes messages from the send channel to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	client := &Client{
		id:   generateClientID(),
		conn: conn,
		hub:  hub,
		send: make(chan []byte, 256),
	}
	hub.addClient(client)

	// Send welcome message
	welcome := Message{
		Type:      "welcome",
		Data:      fmt.Sprintf("Welcome! Your client ID is %s", client.id),
		From:      "server",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(welcome)
	client.send <- data

	go client.writePump()
	client.readPump()
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleMetrics(hub *Hub, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connections":     hub.connectionCount(),
		"messages_total":  hub.getMessageCount(),
		"goroutines":      runtime.NumGoroutine(),
		"redis_enabled":   hub.redisEnabled,
	})
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	hub := newHub()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		handleMetrics(hub, w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting WebSocket server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
