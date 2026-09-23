package docs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestServeResolvesCleanPageURL(t *testing.T) {
	root := t.TempDir()
	page := filepath.Join(root, "features", "virtual-models", "index.html")
	if err := os.MkdirAll(filepath.Dir(page), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(page, []byte("<html><body>virtual models</body></html>"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := New(root, "/")
	req := httptest.NewRequest(http.MethodGet, "/docs/features/virtual-models", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	if err := h.Serve(c); err != nil {
		t.Fatalf("Serve() returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "virtual models") {
		t.Fatalf("expected page body, got %q", rec.Body.String())
	}
}

func TestServeUsesEmbeddedFallback(t *testing.T) {
	h := New("", "/")
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	if err := h.Serve(c); err != nil {
		t.Fatalf("Serve() returned error: %v", err)
	}

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "NEXUS AI Gateway") {
		t.Fatalf("expected embedded fallback page, got status %d and body %q", rec.Code, rec.Body.String())
	}
}
