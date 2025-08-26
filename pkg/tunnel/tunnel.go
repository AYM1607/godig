package tunnel

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/yamux"

	"github.com/AYM1607/godig/pkg/auth"
	"github.com/AYM1607/godig/types"
)

type TunnelClient struct {
	serverAddr string
	localAddr  string
	apiKey     string
	session    *yamux.Session
	conn       net.Conn

	TunnelID string
	Bearer   string
}

func NewTunnelClient(serverAddr, localAddr, apiKey string, clientConfig types.TunnelClientConfig) (*TunnelClient, error) {
	tunnelConfig, err := loadTunnelConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load tunnel config: %w", err)
	}

	if tunnelConfig == nil {
		bearer, err := auth.GenerateString(20)
		if err != nil {
			return nil, fmt.Errorf("failed to generate bearer token: %w", err)
		}

		id, err := auth.GenerateString(5)
		if err != nil {
			return nil, fmt.Errorf("failed to generate tunnel ID: %w", err)
		}

		id = strings.ToLower(id)

		tunnelConfig = &types.TunnelConfig{
			TunnelID: id,
			Bearer:   bearer,
		}

		if clientConfig.PersistConfig {
			if err := saveTunnelConfig(tunnelConfig); err != nil {
				return nil, fmt.Errorf("failed to save tunnel config: %w", err)
			}
		}
	}

	return &TunnelClient{
		Bearer:   tunnelConfig.Bearer,
		TunnelID: tunnelConfig.TunnelID,

		serverAddr: serverAddr,
		localAddr:  localAddr,
		apiKey:     apiKey,
	}, nil
}

// sleepUntilOrCancelled sleeps for the given duration or returns early if the context is cancelled.
// Returns true if the context was cancelled, false if the sleep completed normally.
func sleepUntilOrCancelled(ctx context.Context, duration time.Duration) bool {
	select {
	case <-ctx.Done():
		return true
	case <-time.After(duration):
		return false
	}
}

func (tc *TunnelClient) Run(ctx context.Context) {

	hm := types.HandshakeMessage{
		TunnelID: tc.TunnelID,
		APIKey:   tc.apiKey,
		Bearer:   tc.Bearer,
	}

	// TODO: Try to get the message from the persisted file.
	// TODO: Exponential backoffs for retries.
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping tunnel client")
			return
		default:
		}

		log.Printf("Attempting to connect to tunnel server at %s", tc.serverAddr)

		if err := tc.connect(hm); err != nil {
			log.Printf("Failed to connect: %v", err)
			log.Println("Retrying in 5 seconds...")

			if sleepUntilOrCancelled(ctx, 5*time.Second) {
				return
			}
			continue
		}

		// TODO: Once a connection is accepted, persist it to a file in the current directory.

		// Start handling streams
		tc.start(ctx)

		// Connection lost, cleanup and retry
		if tc.session != nil {
			tc.session.Close()
		}
		if tc.conn != nil {
			tc.conn.Close()
		}

		log.Println("Connection lost. Reconnecting in 5 seconds...")

		if sleepUntilOrCancelled(ctx, 5*time.Second) {
			return
		}
	}
}

// TODO: Pass the handshake message as a parameter.
func (tc *TunnelClient) connect(hm types.HandshakeMessage) error {
	// Connect to tunnel server
	conn, err := net.Dial("tcp", tc.serverAddr)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(hm); err != nil {
		conn.Close()
		return err
	}

	// Wait for acknowledgment
	decoder := json.NewDecoder(conn)
	var response map[string]string
	if err := decoder.Decode(&response); err != nil {
		conn.Close()
		return err
	}

	if response["status"] != "ok" {
		conn.Close()
		return fmt.Errorf("handshake failed: %v", response)
	}

	// Create yamux session
	session, err := yamux.Client(conn, yamux.DefaultConfig())
	if err != nil {
		conn.Close()
		return err
	}

	tc.conn = conn
	tc.session = session

	log.Printf("Connected to tunnel server. Public URL: http://%s.%s", tc.TunnelID, tc.serverAddr)
	return nil
}

func (tc *TunnelClient) start(ctx context.Context) {
	for {
		stream, err := tc.session.AcceptStreamWithContext(ctx)
		if err != nil {
			log.Printf("Failed to accept stream: %v", err)
			break
		}

		// Handle each stream in a goroutine
		go tc.handleStream(ctx, stream)
	}
}

func (tc *TunnelClient) handleStream(ctx context.Context, stream net.Conn) {
	defer stream.Close()

	// Set up a goroutine to close the stream if context is cancelled
	go func() {
		<-ctx.Done()
		stream.Close()
	}()

	// Set timeout ONLY for reading the initial HTTP request
	stream.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Read HTTP request from the stream
	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		log.Printf("Failed to read request from stream: %v", err)
		return
	}

	// IMPORTANT: Clear deadline immediately after reading request
	// This allows long-running connections (SSE, WebSocket, etc.)
	stream.SetReadDeadline(time.Time{})

	log.Printf("Handling request: %s %s", req.Method, req.URL.Path)

	// Connect to local service
	localConn, err := net.Dial("tcp", tc.localAddr)
	if err != nil {
		log.Printf("Failed to connect to local service: %v", err)
		// Send error response
		errorResp := "HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n"
		stream.Write([]byte(errorResp))
		return
	}
	defer localConn.Close()

	// Forward the request to local service
	if err := req.Write(localConn); err != nil {
		log.Printf("Failed to write request to local service: %v", err)
		return
	}

	// Copy response from local service back to tunnel stream
	// This handles both regular HTTP responses and streaming responses (SSE, chunked, etc.)
	go func() {
		defer stream.Close()
		defer localConn.Close()
		io.Copy(stream, localConn)
	}()

	// Copy any remaining request data (for uploads, etc.)
	io.Copy(localConn, stream)
}
