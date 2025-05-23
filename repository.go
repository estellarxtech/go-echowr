// Code generated by ifacemaker; DO NOT EDIT.

package server

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ServerRepo ...
type ServerRepo interface {
	// NewContext creates a new Echo context
	NewContext(req *http.Request, w http.ResponseWriter) Context
	// RegisterRouters registers multiple routers with the specified group and middlewares
	RegisterRouters(group Kind, routers *RegisterRouters, middlewares ...MiddlewareFunc) error
	// Start starts the server
	Start()
	// GetEcho returns the Echo instance
	GetEcho() *echo.Echo
	// GetRouters returns all registered routes
	GetRouters() []*Route
	// Close closes the server
	Close() error
	// Shutdown gracefully shuts down the server
	Shutdown(ctx context.Context) error
	// GracefulShutdown shuts down the server with a timeout
	GracefulShutdown() error
}
