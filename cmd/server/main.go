package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/yamux"

	"github.com/AYM1607/godig/pkg/auth"
	"github.com/AYM1607/godig/pkg/headers"
	"github.com/AYM1607/godig/types"
)

type TunnelServer struct {
	clients map[string]*ClientSession
	mutex   sync.RWMutex
	apiKey  string
}

type ClientSession struct {
	ID      string
	Session *yamux.Session
	Conn    net.Conn
	Bearer  string
}

func main() {
	server := NewTunnelServer()

	go func() {
		listener, err := net.Listen("tcp", ":8080")
		if err != nil {
			log.Fatal("Failed to start tunnel listener:", err)
		}
		defer listener.Close()

		log.Println("Tunnel server listening on :8080")

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v", err)
				continue
			}

			go server.handleTunnelConnection(conn)
		}
	}()

	http.HandleFunc("/", server.ServeHTTP)

	log.Println("HTTP server listening on :8081")
	log.Printf("Access tunnels at: https://{tunnel-id}.%s:8081\n", getHost())
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func NewTunnelServer() *TunnelServer {
	key, err := auth.GetServerKey()
	if err != nil {
		log.Fatalln(err)
	}
	return &TunnelServer{
		clients: make(map[string]*ClientSession),
		apiKey:  key,
	}
}

func (ts *TunnelServer) handleTunnelConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	decoder := json.NewDecoder(conn)
	var handshake types.HandshakeMessage
	if err := decoder.Decode(&handshake); err != nil {
		log.Printf("Failed to read handshake: %v", err)
		return
	}

	if handshake.APIKey != ts.apiKey {
		log.Printf("Invalid API Key")
		return
	}

	// TODO: Validate max length.
	if handshake.TunnelID == "" {
		log.Printf("Invalid tunnel ID in handshake")
		return
	}

	// TODO: Validate minimum security.
	if handshake.Bearer == "" {
		log.Printf("Invalid bearer in handshake")
		return
	}

	log.Printf("Client connecting with tunnel ID: %s", handshake.TunnelID)

	// Send acknowledgment
	response := map[string]string{"status": "ok"}
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		log.Printf("Failed to send handshake response: %v", err)
		return
	}

	// Clear read deadline. TODO: Understand why this is needed.
	conn.SetReadDeadline(time.Time{})

	session, err := yamux.Server(conn, yamux.DefaultConfig())
	if err != nil {
		log.Printf("Failed to create yamux session: %v", err)
		return
	}

	// Register client
	clientSession := &ClientSession{
		ID:      handshake.TunnelID,
		Session: session,
		Conn:    conn,
		Bearer:  handshake.Bearer,
	}

	ts.registerClient(clientSession)
	defer ts.unregisterClient(handshake.TunnelID)

	log.Printf("Tunnel established for %s.tunnel.local", handshake.TunnelID)

	// Keep connection alive until client disconnects
	<-session.CloseChan()
	log.Printf("Client %s disconnected", handshake.TunnelID)
}

func (ts *TunnelServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract tunnel ID from subdomain
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		http.Error(w, "Invalid subdomain", http.StatusBadRequest)
		return
	}

	tunnelID := parts[0]
	if tunnelID == "" {
		http.Error(w, "Missing tunnel ID", http.StatusBadRequest)
		return
	}

	client := ts.getClient(tunnelID)
	if client == nil {
		http.Error(w, "Tunnel not found or not connected", http.StatusServiceUnavailable)
		return
	}

	token := getBearerToken(r)
	if token != client.Bearer {
		http.Error(w, "Auth failed", http.StatusBadGateway)
		return
	}

	stream, err := client.Session.Open()
	if err != nil {
		log.Printf("Failed to open stream for %s: %v", tunnelID, err)
		http.Error(w, "Failed to open tunnel stream", http.StatusBadGateway)
		return
	}
	defer stream.Close()

	// TODO: This might break long SSE.
	// Set timeout for the stream
	stream.SetDeadline(time.Now().Add(30 * time.Second))

	// Forward the HTTP request to the client
	if err := r.Write(stream); err != nil {
		log.Printf("Failed to write request to stream: %v", err)
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}

	// Read the HTTP response from the client
	resp, err := http.ReadResponse(bufio.NewReader(stream), r)
	if err != nil {
		log.Printf("Failed to read response from stream: %v", err)
		http.Error(w, "Failed to read response", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Get client IP for proxy headers
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	// Clean up request headers and add proxy headers
	headers.RemoveHopByHopHeaders(r.Header)
	headers.AddProxyHeaders(r, clientIP)

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Send headers before streaming the response to handle SSE gracefully.
	w.WriteHeader(resp.StatusCode)

	if isStreamingResponse(resp) {
		log.Printf("Handling streaming response for %s", tunnelID)
		if err := ts.handleStreamingResponse(w, resp); err != nil {
			log.Printf("Error handling streaming response: %v", err)
		}
	} else {
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Printf("Error copying response body: %v", err)
		}
	}

}

func (ts *TunnelServer) registerClient(client *ClientSession) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	// Close existing session if any.
	if existing, exists := ts.clients[client.ID]; exists {
		log.Printf("Replacing existing session for tunnel ID: %s", client.ID)
		// TODO: Handle these errors.
		existing.Session.Close()
		existing.Conn.Close()
	}

	ts.clients[client.ID] = client
}

func (ts *TunnelServer) unregisterClient(tunnelID string) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	delete(ts.clients, tunnelID)
}

func (ts *TunnelServer) getClient(tunnelID string) *ClientSession {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return ts.clients[tunnelID]
}

func getHost() string {
	host := os.Getenv("GODIG_HOST")
	if host == "" {
		host = "localhost"
	}
	return host
}

func getBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if it starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	// Extract the token part
	return strings.TrimPrefix(authHeader, "Bearer ")
}
