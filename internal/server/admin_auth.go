package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nexusrun/nexus_aigateway/internal/adminauth"
)

type adminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func registerAdminAuth(e *echo.Echo, auth *adminauth.Service) {
	e.POST("/admin/auth/login", func(c *echo.Context) error {
		var req adminLoginRequest
		if err := c.Bind(&req); err != nil || strings.TrimSpace(req.Username) == "" || req.Password == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		}
		token, err := auth.Login(c.Request().Context(), req.Username, req.Password)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		}
		auth.SetSession(c.Response(), token, c.Request().TLS != nil)
		return c.JSON(http.StatusOK, map[string]bool{"authenticated": true})
	})
	e.POST("/admin/auth/logout", func(c *echo.Context) error {
		auth.ClearSession(c.Response(), c.Request().TLS != nil)
		return c.JSON(http.StatusOK, map[string]bool{"authenticated": false})
	})
	e.GET("/admin/auth/session", func(c *echo.Context) error {
		identity, err := auth.AuthenticateRequest(c.Request().Context(), c.Request())
		return c.JSON(http.StatusOK, map[string]bool{"authenticated": err == nil && identity != nil})
	})
}
