package headers

import (
	"net/http"
	"strings"
)

var hopByHopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Te":                  true,
	"Trailer":             true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
}

func RemoveHopByHopHeaders(header http.Header) {
	for name := range hopByHopHeaders {
		header.Del(name)
	}

	// Also remove headers listed in Connection header
	if connectionHeader := header.Get("Connection"); connectionHeader != "" {
		for _, h := range strings.Split(connectionHeader, ",") {
			header.Del(strings.TrimSpace(h))
		}
	}
}

func AddProxyHeaders(req *http.Request, clientIP string) {
	// Add X-Forwarded-For
	if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
		req.Header.Set("X-Forwarded-For", prior+", "+clientIP)
	} else {
		req.Header.Set("X-Forwarded-For", clientIP)
	}

	// Add X-Forwarded-Proto
	if req.TLS != nil {
		req.Header.Set("X-Forwarded-Proto", "https")
	} else {
		req.Header.Set("X-Forwarded-Proto", "http")
	}

	// Add X-Forwarded-Host
	if req.Header.Get("X-Forwarded-Host") == "" {
		req.Header.Set("X-Forwarded-Host", req.Host)
	}
}
