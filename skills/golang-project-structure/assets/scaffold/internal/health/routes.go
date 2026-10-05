package health

import "github.com/labstack/echo/v4"

// Register mounts GET /health on g (the public group, so /api/health).
func Register(g *echo.Group, deps Deps) {
	h := &handlers{svc: deps.Service}
	g.GET("/health", h.handleGet)
}
