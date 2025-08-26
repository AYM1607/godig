package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

func isStreamingResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "text/event-stream") ||
		resp.Header.Get("Transfer-Encoding") == "chunked" ||
		resp.Header.Get("Connection") == "keep-alive"
}

// handleStreamingResponse sends keep-alive comments on top of the regular
// content to prevent connections from being closed.
func (ts *TunnelServer) handleStreamingResponse(w http.ResponseWriter, resp *http.Response, stream net.Conn) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("response writer doesn't support flushing")
	}

	// Mutex to protect concurrent writes to the response writer
	var writeMutex sync.Mutex

	// 25 seconds is under most browser/client/proxy timeouts.
	keepaliveTicker := time.NewTicker(25 * time.Second)
	defer keepaliveTicker.Stop()

	done := make(chan error, 1)

	go func() {
		buf := make([]byte, 32*1024) // 32KB buffer for better performance
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				writeMutex.Lock()
				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
					writeMutex.Unlock()
					done <- writeErr
					return
				}
				flusher.Flush()
				writeMutex.Unlock()
			}
			if err != nil {
				if err == io.EOF {
					done <- nil
				} else {
					done <- err
				}
				return
			}
		}
	}()

	for {
		select {
		case <-keepaliveTicker.C:
			// Send SSE keepalive comment (ignored by browsers but keeps connection alive)
			if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
				writeMutex.Lock()
				_, err := w.Write([]byte(": keepalive\n\n"))
				if err != nil {
					return err
				}
				// Extend the deadline to avoid closures by the multiplexer.
				stream.SetDeadline(time.Now().Add(60 * time.Second))
				flusher.Flush()
				writeMutex.Unlock()
			}
		case err := <-done:
			return err
		}
	}
}
