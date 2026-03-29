package activity

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPost_202Pending(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":     false,
			"pending":     true,
			"approvalUrl": "https://example.com/admin/devices?approve=1",
			"error":       "pending",
		})
	}))
	defer ts.Close()

	c := &Client{BaseURL: ts.URL, Token: "tok"}
	err := c.Post(context.Background(), ReportRequest{
		GeneratedHashKey: "abc",
		ProcessName:      "code",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var p *PendingApprovalError
	if !errors.As(err, &p) {
		t.Fatalf("expected PendingApprovalError, got %T %v", err, err)
	}
	if p.ApprovalURL == "" {
		t.Fatal("empty approval url")
	}
}
