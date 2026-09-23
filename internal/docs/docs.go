// Package docs serves the statically exported product documentation.
package docs

import (
	"bytes"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/nexusrun/nexus_aigateway/config"
)

// static contains a small fallback page so a source checkout remains
// buildable. Production images replace it with the Mintlify export in /app/docs.
//
//go:embed all:static
var static embed.FS

// Handler serves the product documentation below /docs.
type Handler struct {
	content    fs.FS
	fileServer http.Handler
	publicPath string
}

// New creates a documentation handler. If root is empty or unavailable, the
// embedded fallback page is served instead.
func New(root, basePath string) *Handler {
	content, err := fs.Sub(static, "static")
	if err != nil {
		panic("embedded documentation assets are missing")
	}

	if root = strings.TrimSpace(root); root != "" {
		info, statErr := os.Stat(root)
		if statErr == nil && info.IsDir() {
			content = os.DirFS(root)
		} else {
			slog.Warn("documentation directory is unavailable; using embedded fallback", "path", root, "error", statErr)
		}
	}

	return &Handler{
		content:    content,
		fileServer: http.FileServer(http.FS(content)),
		publicPath: config.JoinBasePath(basePath, "/docs"),
	}
}

// Serve handles /docs and /docs/* requests. Clean page URLs are resolved to
// their exported index.html files so callers do not need a trailing slash.
func (h *Handler) Serve(c *echo.Context) error {
	relative := strings.TrimPrefix(c.Request().URL.Path, "/docs")
	relative = strings.TrimPrefix(relative, "/")
	filePath := h.resolve(relative)
	if filePath == "" {
		http.NotFound(c.Response(), c.Request())
		return nil
	}

	if strings.HasSuffix(filePath, ".html") {
		return h.serveHTML(c, filePath)
	}

	request := c.Request().Clone(c.Request().Context())
	urlCopy := *request.URL
	urlCopy.Path = "/" + filePath
	urlCopy.RawPath = ""
	request.URL = &urlCopy
	h.fileServer.ServeHTTP(c.Response(), request)
	return nil
}

func (h *Handler) resolve(relative string) string {
	if relative == "" {
		relative = "index.html"
	}
	relative = path.Clean(relative)
	if relative == "." || strings.HasPrefix(relative, "../") || relative == ".." {
		return ""
	}

	if info, err := fs.Stat(h.content, relative); err == nil && !info.IsDir() {
		return relative
	}
	for _, candidate := range []string{path.Join(relative, "index.html"), relative + ".html"} {
		if info, err := fs.Stat(h.content, candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func (h *Handler) serveHTML(c *echo.Context, filePath string) error {
	data, err := fs.ReadFile(h.content, filePath)
	if err != nil {
		http.NotFound(c.Response(), c.Request())
		return nil
	}
	data = bytes.ReplaceAll(data, []byte(`"/docs`), []byte(`"`+h.publicPath))
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().Header().Set("Cache-Control", "no-cache")
	_, err = c.Response().Write(data)
	return err
}
