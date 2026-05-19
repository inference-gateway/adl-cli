package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_FetchByID(t *testing.T) {
	var lastPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastPath = r.URL.Path
		switch r.URL.Path {
		case "/skills/data-analysis.md":
			w.Header().Set("Content-Type", "text/markdown")
			_, _ = w.Write([]byte("---\nname: data-analysis\ndescription: x\n---\nbody\n"))
		case "/skills/data-analysis/1.2.3.md":
			w.Header().Set("Content-Type", "text/markdown")
			_, _ = w.Write([]byte("---\nname: data-analysis\ndescription: pinned\n---\nbody\n"))
		case "/skills/missing.md":
			http.NotFound(w, r)
		default:
			http.Error(w, "unexpected", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL + "/skills/")

	body, err := client.FetchByID(context.Background(), "data-analysis", "")
	if err != nil {
		t.Fatalf("FetchByID without version: %v", err)
	}
	if !strings.Contains(string(body), "name: data-analysis") {
		t.Errorf("unexpected body: %q", body)
	}
	if lastPath != "/skills/data-analysis.md" {
		t.Errorf("expected GET /skills/data-analysis.md, got %s", lastPath)
	}

	body, err = client.FetchByID(context.Background(), "data-analysis", "1.2.3")
	if err != nil {
		t.Fatalf("FetchByID with version: %v", err)
	}
	if !strings.Contains(string(body), "pinned") {
		t.Errorf("expected pinned body, got %q", body)
	}
	if lastPath != "/skills/data-analysis/1.2.3.md" {
		t.Errorf("expected versioned path, got %s", lastPath)
	}

	if _, err := client.FetchByID(context.Background(), "missing", ""); err == nil {
		t.Error("expected 404 error, got nil")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %v", err)
	}
}

func TestNewClient_DefaultBaseURL(t *testing.T) {
	c := NewClient("")
	if c.BaseURL != DefaultBaseURL {
		t.Errorf("default BaseURL = %q, want %q", c.BaseURL, DefaultBaseURL)
	}

	c = NewClient("https://example.com/skills")
	if c.BaseURL != "https://example.com/skills/" {
		t.Errorf("trailing slash should be added, got %q", c.BaseURL)
	}
}
