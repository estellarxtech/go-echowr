//go:generate ifacemaker -f $GOFILE -s Server -i ServerRepo -p server -o repository.go
//go:generate mockgen -source=repository.go -package=${GOPACKAGE} -destination=${GOPACKAGE}_mock.go

package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// Kind represents the type of router group
type Kind int

const (
	ROOT Kind = iota
	V1
	V2
	V3
	DEV
	API
	DOCS
)

func (k Kind) String() string {
	return [...]string{
		"root",
		"v1",
		"v2",
		"v3",
		"dev",
		"api",
		"docs",
	}[k]
}

// RegisterRouter defines a single router with a path and methods
type RegisterRouter struct {
	Path    string
	Methods map[string]HandlerFunc
}

// RegisterRouters holds multiple routers with a fixed path prefix
type RegisterRouters struct {
	PathFixed string
	Routers   []RegisterRouter
}

// NewRouters creates a new instance of RegisterRouters
func NewRouters() *RegisterRouters {
	return &RegisterRouters{}
}

// AddRouter adds a new router to the list
func (r *RegisterRouters) AddRouter(path string, methods map[string]HandlerFunc) {
	r.Routers = append(r.Routers, RegisterRouter{
		Path:    path,
		Methods: methods,
	})
}

// AddRouterFx adds a new router with a fixed path prefix
func (r *RegisterRouters) AddRouterFx(params string, methods map[string]HandlerFunc) {
	path := strings.TrimSpace(params)
	if len(path) > 0 {
		path = r.PathFixed + path
	} else {
		path = r.PathFixed
	}

	r.Routers = append(r.Routers, RegisterRouter{
		Path:    path,
		Methods: methods,
	})
}

// GetAllRouters returns all registered routers
func (r *RegisterRouters) GetAllRouters() []RegisterRouter {
	return r.Routers
}

// GetRouters returns routers matching the specified path
func (r *RegisterRouters) GetRouters(path string) []RegisterRouter {
	var routers []RegisterRouter
	for _, router := range r.Routers {
		if router.Path == path {
			routers = append(routers, router)
		}
	}
	return routers
}

// GetRoutersFx returns routers containing the fixed path prefix
func (r *RegisterRouters) GetRoutersFx() []RegisterRouter {
	var routers []RegisterRouter
	for _, router := range r.Routers {
		if strings.Contains(router.Path, r.PathFixed) {
			routers = append(routers, router)
		}
	}
	return routers
}

// SetPathFixed sets the fixed path prefix
func (r *RegisterRouters) SetPathFixed(path string) {
	r.PathFixed = path
}

type Methods map[string]HandlerFunc
type HandlerFunc = echo.HandlerFunc
type MiddlewareFunc = echo.MiddlewareFunc
type Context = echo.Context
type Route = echo.Route

// Server represents the HTTP server
type Server struct {
	port   string
	host   string
	echo   *echo.Echo
	params *ServerParams
}

// NewServer creates a new server instance with the given options
func NewServer(opts ...Options) (*Server, error) {
	params, err := newServerParams(opts...)
	if err != nil {
		return nil, err
	}

	e := echo.New()

	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())
	// e.Use(middleware.CORS())

	e.HideBanner = true

	return &Server{
		echo:   e,
		port:   params.GetPort(),
		host:   params.GetHost(),
		params: params,
	}, nil
}

func (s *Server) Use(middleware MiddlewareFunc) {
	s.echo.Use(middleware)
}

func (s *Server) Uses(middlewares ...MiddlewareFunc) {
	s.echo.Use(middlewares...)
}

// NewContext creates a new Echo context
func (s *Server) NewContext(req *http.Request, w http.ResponseWriter) Context {
	return s.echo.NewContext(req, w)
}

// RegisterRouters registers multiple routers with the specified group and middlewares
func (s *Server) RegisterRouters(group Kind, routers *RegisterRouters, middlewares ...MiddlewareFunc) error {
	var grp any

	switch group {
	case ROOT:
		grp = s.echo
	case V1, V2, V3, DEV, API, DOCS:
		grp = s.echo.Group(group.String())
	default:
		return fmt.Errorf("invalid group type")
	}

	return s.registerRouters(grp, routers, middlewares...)
}

// registerRouters registers routers to the given Echo group or instance
func (s *Server) registerRouters(engine any, routers *RegisterRouters, middlewares ...MiddlewareFunc) error {
	for _, middleware := range middlewares {
		switch e := engine.(type) {
		case *echo.Group:
			e.Use(middleware)
		case *echo.Echo:
			e.Use(middleware)
		}
	}

	for _, methods := range routers.GetAllRouters() {
		for method, handler := range methods.Methods {
			if err := s.registerMethod(engine, method, methods.Path, handler); err != nil {
				return err
			}
		}
	}

	return nil
}

// registerMethod registers a single method to the Echo instance
func (s *Server) registerMethod(engine any, method, path string, handler echo.HandlerFunc) error {
	switch e := engine.(type) {
	case *echo.Group:
		switch method {
		case http.MethodGet:
			e.GET(path, handler)
		case http.MethodPost:
			e.POST(path, handler)
		case http.MethodPut:
			e.PUT(path, handler)
		case http.MethodDelete:
			e.DELETE(path, handler)
		case http.MethodPatch:
			e.PATCH(path, handler)
		case http.MethodHead:
			e.HEAD(path, handler)
		case http.MethodConnect:
			e.CONNECT(path, handler)
		case http.MethodOptions:
			e.OPTIONS(path, handler)
		case http.MethodTrace:
			e.TRACE(path, handler)
		default:
			return fmt.Errorf("unsupported method: %s", method)
		}

	case *echo.Echo:
		switch method {
		case http.MethodGet:
			e.GET(path, handler)
		case http.MethodPost:
			e.POST(path, handler)
		case http.MethodPut:
			e.PUT(path, handler)
		case http.MethodDelete:
			e.DELETE(path, handler)
		case http.MethodPatch:
			e.PATCH(path, handler)
		case http.MethodHead:
			e.HEAD(path, handler)
		case http.MethodConnect:
			e.CONNECT(path, handler)
		case http.MethodOptions:
			e.OPTIONS(path, handler)
		case http.MethodTrace:
			e.TRACE(path, handler)
		default:
			return fmt.Errorf("unsupported method: %s", method)
		}
	default:
		return fmt.Errorf("engine type not supported")
	}

	return nil
}

// Start starts the server
func (s *Server) Start() {
	host := fmt.Sprintf("%s:%s", s.host, s.port)
	if len(s.port) == 0 {
		host = s.host
	}

	go func() {
		if err := s.echo.Start(host); err != nil && err != http.ErrServerClosed {
			s.echo.Logger.Fatal(err)
		}
	}()
}

// GetEcho returns the Echo instance
func (s *Server) GetEcho() *echo.Echo {
	return s.echo
}

// GetRouters returns all registered routes
func (s *Server) GetRouters() []*Route {
	return s.echo.Routes()
}

// Close closes the server
func (s *Server) Close() error {
	return s.echo.Close()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

// GracefulShutdown shuts down the server with a timeout
func (s *Server) GracefulShutdown() error {
	return s.gracefulShutdown()
}

func (s *Server) gracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.Shutdown(ctx)
}
