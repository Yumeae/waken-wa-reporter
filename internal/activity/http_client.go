package activity

import (
	"net/http"
	"time"
)

const reporterHTTPTimeout = 30 * time.Second

// HTTPClientBypassProxy returns a client that ignores HTTP_PROXY / HTTPS_PROXY / NO_PROXY
// (Transport.Proxy is nil). Use when the API is reachable only without a corporate proxy.
func HTTPClientBypassProxy() *http.Client {
	tr, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Client{
			Timeout:   reporterHTTPTimeout,
			Transport: &http.Transport{Proxy: nil},
		}
	}
	cloned := tr.Clone()
	cloned.Proxy = nil
	return &http.Client{
		Timeout:   reporterHTTPTimeout,
		Transport: cloned,
	}
}
