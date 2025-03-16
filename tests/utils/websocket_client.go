package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient represents a client for interacting with WebSocket endpoints
type WebSocketClient struct {
	conn      *websocket.Conn
	url       string
	idCounter int
	mu        sync.Mutex
	doneCh    chan struct{}
}

// WebSocketRequest represents a request to a WebSocket API
type WebSocketRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int         `json:"id"`
}

// WebSocketResponse represents a response from a WebSocket API
type WebSocketResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data,omitempty"`
	} `json:"error,omitempty"`
}

// EventData represents subscription event data
type EventData struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

// NewWebSocketClient creates a new WebSocket client for the specified URL
func NewWebSocketClient(url string) (*WebSocketClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}

	client := &WebSocketClient{
		conn:      conn,
		url:       url,
		idCounter: 1,
		doneCh:    make(chan struct{}),
	}

	return client, nil
}

// Close closes the WebSocket connection
func (c *WebSocketClient) Close() error {
	close(c.doneCh)
	return c.conn.Close()
}

// nextID returns the next request ID
func (c *WebSocketClient) nextID() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.idCounter
	c.idCounter++
	return id
}

// Subscribe subscribes to an event and returns a channel for receiving events
func (c *WebSocketClient) Subscribe(ctx context.Context, query string) (<-chan EventData, error) {
	// Create request
	req := WebSocketRequest{
		JSONRPC: "2.0",
		Method:  "subscribe",
		Params:  map[string]string{"query": query},
		ID:      c.nextID(),
	}

	// Send request
	if err := c.conn.WriteJSON(req); err != nil {
		return nil, fmt.Errorf("failed to send subscription request: %w", err)
	}

	// Wait for subscription response
	var resp WebSocketResponse
	if err := c.conn.ReadJSON(&resp); err != nil {
		return nil, fmt.Errorf("failed to read subscription response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("subscription failed: %s", resp.Error.Message)
	}

	// Create a channel for the subscription events
	eventCh := make(chan EventData, 100)

	// Start a goroutine to receive events
	go c.receiveEvents(eventCh, ctx)

	return eventCh, nil
}

// receiveEvents receives events from the WebSocket connection and sends them to the event channel
func (c *WebSocketClient) receiveEvents(eventCh chan<- EventData, ctx context.Context) {
	defer close(eventCh)

	for {
		select {
		case <-c.doneCh:
			return
		case <-ctx.Done():
			return
		default:
			// Set read deadline to avoid blocking forever
			_ = c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			// Read next message
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				// If the error is a timeout, continue
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				}
				fmt.Printf("Error reading from WebSocket: %v\n", err)
				return
			}

			// Parse message
			var event struct {
				JSONRPC string    `json:"jsonrpc"`
				Method  string    `json:"method"`
				Params  EventData `json:"params"`
			}

			if err := json.Unmarshal(message, &event); err != nil {
				fmt.Printf("Error unmarshaling event: %v\n", err)
				continue
			}

			// Send event to channel
			select {
			case eventCh <- event.Params:
			default:
				fmt.Println("Warning: Event channel is full, dropping event")
			}
		}
	}
}

// Query sends a query to the WebSocket API and returns the response
func (c *WebSocketClient) Query(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	// Create request
	req := WebSocketRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      c.nextID(),
	}

	// Send request
	if err := c.conn.WriteJSON(req); err != nil {
		return nil, fmt.Errorf("failed to send query request: %w", err)
	}

	// Wait for response with a timeout
	respCh := make(chan WebSocketResponse, 1)
	errCh := make(chan error, 1)

	go func() {
		var resp WebSocketResponse
		if err := c.conn.ReadJSON(&resp); err != nil {
			errCh <- fmt.Errorf("failed to read query response: %w", err)
			return
		}
		respCh <- resp
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, fmt.Errorf("query failed: %s", resp.Error.Message)
		}
		return resp.Result, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for query response")
	}
}
