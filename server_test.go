package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer()
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

func TestRegisterRouters(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, rr.GetRouters("/test")[0].Methods[http.MethodGet](c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test passed", rec.Body.String())
	}
}

func TestStartAndShutdown(t *testing.T) {
	server, _ := NewServer()

	go server.Start()
	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	assert.NoError(t, server.Shutdown(ctx))
}

func TestServerClose(t *testing.T) {
	server, _ := NewServer()

	go server.Start()
	time.Sleep(1 * time.Second)

	assert.NoError(t, server.Close())
}

func TestGetEcho(t *testing.T) {
	server, _ := NewServer()

	assert.IsType(t, &echo.Echo{}, server.GetEcho())
}

func TestGracefulShutdown(t *testing.T) {
	server, _ := NewServer()

	go server.Start()
	time.Sleep(1 * time.Second)

	assert.NoError(t, server.gracefulShutdown())
}

func TestGetRouters(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	routers := server.GetRouters()
	assert.Len(t, routers, 1)
}

func TestRegisterRoutes(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	routers := server.GetRouters()
	assert.Len(t, routers, 1)
}

func TestRegisterRoutesAndValidEndpoint(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, rr.GetRouters("/test")[0].Methods[http.MethodGet](c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test passed", rec.Body.String())
	}
}

func TestRouterNotFound(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	rr2 := NewRouters()
	rr2.AddRouter("/test2", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr2)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRouterMethodNotAllowed(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRouterFixedPath(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test passed", rec.Body.String())
}

func TestRouterFixedPathNotFound(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRouterFixedPathMethodNotAllowed(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRouterFixedPathAndValidEndpoint(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test passed", rec.Body.String())
}

func TestRouterFixedPathAndValidEndpointWithParams(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test/:id", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, c.Param("id"))
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/test/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "123", rec.Body.String())
}

func TestRouterFixedPathAndValidEndpointWithParamsAndQuery(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test/:id", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, c.QueryParam("q"))
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/test/123?q=abc", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "abc", rec.Body.String())
}

func TestRouterFixedPathAndValidEndpointWithParamsAndQueryAndBody(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test/:id", map[string]HandlerFunc{
		http.MethodPost: func(c Context) error {
			type body struct {
				Q string `json:"q"`
			}

			b := new(body)
			if err := c.Bind(b); err != nil {
				return err
			}

			return c.String(http.StatusOK, b.Q)
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()

	bodyContent := `{"q":"abc"}`
	req := httptest.NewRequest(http.MethodPost, "/api/test/123", bytes.NewReader([]byte(bodyContent)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "abc", rec.Body.String())
}

func TestRoutesWithGroup(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(V1, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test passed", rec.Body.String())
}

func TestRoutesWithGroupAndValidEndpoint(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(V1, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test passed", rec.Body.String())
}

func TestRoutesWithGroupAndValidEndpointWithParams(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test/:id", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, c.Param("id"))
		},
	})

	_ = server.RegisterRouters(V1, rr)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/v1/test/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "123", rec.Body.String())
}

func TestRoutesFixedPathWithMultiGroups(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(V1, rr)

	rr2 := NewRouters()
	rr2.SetPathFixed("/api")
	rr2.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(V2, rr2)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/v2/api/test", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test passed", rec.Body.String())

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "test passed", rec2.Body.String())
}

func TestRoutesFixedPathWithMultiGroupsMethodGetNotAllowed(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(V1, rr)

	rr2 := NewRouters()
	rr2.SetPathFixed("/api")
	rr2.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(V2, rr2)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodPost, "/v2/api/test", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, http.StatusMethodNotAllowed, rec2.Code)
}

func TestRoutesFixedPathWithMultiGroupsMethodGetParams(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test/:id", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, c.Param("id"))
		},
	})

	_ = server.RegisterRouters(V1, rr)

	rr2 := NewRouters()
	rr2.SetPathFixed("/api")
	rr2.AddRouterFx("/test/:id", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, c.Param("id"))
		},
	})

	_ = server.RegisterRouters(V2, rr2)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/test/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/v2/api/test/123", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "123", rec.Body.String())

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "123", rec2.Body.String())
}

func TestRoutesFixedPathWithMultiGroupsMethodGetParamsAndQuery(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test/:id", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, c.QueryParam("q"))
		},
	})

	_ = server.RegisterRouters(V1, rr)

	rr2 := NewRouters()
	rr2.SetPathFixed("/api")
	rr2.AddRouterFx("/test/:id", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, c.QueryParam("q"))
		},
	})

	_ = server.RegisterRouters(V2, rr2)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/test/123?q=abc", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/v2/api/test/123?q=abc", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "abc", rec.Body.String())

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "abc", rec2.Body.String())
}

func TestRoutesFixedPathWithMultiGroupsMethodGetPassEmptyValue(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test/:id", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, c.Param("id"))
		},
	})

	_ = server.RegisterRouters(V1, rr)

	rr2 := NewRouters()
	rr2.SetPathFixed("/api")
	rr2.AddRouterFx("/test/:id", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, c.Param("id"))
		},
	})

	_ = server.RegisterRouters(V2, rr2)

	e := server.GetEcho()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/test/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/v2/api/test/", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"message\":\"Not Found\"")

	assert.Equal(t, http.StatusNotFound, rec2.Code)
	assert.Contains(t, rec2.Body.String(), "\"message\":\"Not Found\"")
}

func TestHTTPMethodsGroup(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.SetPathFixed("/api")

	methods := map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "GET method")
		},
		http.MethodPost: func(c Context) error {
			return c.String(http.StatusOK, "POST method")
		},
		http.MethodPut: func(c Context) error {
			return c.String(http.StatusOK, "PUT method")
		},
		http.MethodDelete: func(c Context) error {
			return c.String(http.StatusOK, "DELETE method")
		},
		http.MethodPatch: func(c Context) error {
			return c.String(http.StatusOK, "PATCH method")
		},
		http.MethodHead: func(c Context) error {
			return c.NoContent(http.StatusOK)
		},
		http.MethodConnect: func(c Context) error {
			return c.NoContent(http.StatusOK)
		},
		http.MethodOptions: func(c Context) error {
			return c.NoContent(http.StatusNoContent)
		},
		http.MethodTrace: func(c Context) error {
			return c.String(http.StatusOK, "TRACE method")
		},
	}

	rr.AddRouterFx("/test", methods)
	_ = server.RegisterRouters(V1, rr)

	e := server.GetEcho()

	tests := []struct {
		method       string
		expectedBody string
		expectedCode int
	}{
		{http.MethodGet, "GET method", http.StatusOK},
		{http.MethodPost, "POST method", http.StatusOK},
		{http.MethodPut, "PUT method", http.StatusOK},
		{http.MethodDelete, "DELETE method", http.StatusOK},
		{http.MethodPatch, "PATCH method", http.StatusOK},
		{http.MethodHead, "", http.StatusOK},
		{http.MethodConnect, "", http.StatusOK},
		{http.MethodOptions, "", http.StatusNoContent},
		{http.MethodTrace, "TRACE method", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/v1/api/test", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
			if tt.method != http.MethodHead && tt.method != http.MethodConnect && tt.method != http.MethodOptions {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestHTTPMethodsWithoutGroup(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()

	methods := map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "GET method")
		},
		http.MethodPost: func(c Context) error {
			return c.String(http.StatusOK, "POST method")
		},
		http.MethodPut: func(c Context) error {
			return c.String(http.StatusOK, "PUT method")
		},
		http.MethodDelete: func(c Context) error {
			return c.String(http.StatusOK, "DELETE method")
		},
		http.MethodPatch: func(c Context) error {
			return c.String(http.StatusOK, "PATCH method")
		},
		http.MethodHead: func(c Context) error {
			return c.NoContent(http.StatusOK)
		},
		http.MethodConnect: func(c Context) error {
			return c.NoContent(http.StatusOK)
		},
		http.MethodOptions: func(c Context) error {
			return c.NoContent(http.StatusNoContent)
		},
		http.MethodTrace: func(c Context) error {
			return c.String(http.StatusOK, "TRACE method")
		},
	}

	rr.AddRouter("/api/test", methods)
	_ = server.RegisterRouters(ROOT, rr)

	e := server.GetEcho()

	tests := []struct {
		method       string
		expectedBody string
		expectedCode int
	}{
		{http.MethodGet, "GET method", http.StatusOK},
		{http.MethodPost, "POST method", http.StatusOK},
		{http.MethodPut, "PUT method", http.StatusOK},
		{http.MethodDelete, "DELETE method", http.StatusOK},
		{http.MethodPatch, "PATCH method", http.StatusOK},
		{http.MethodHead, "", http.StatusOK},
		{http.MethodConnect, "", http.StatusOK},
		{http.MethodOptions, "", http.StatusNoContent},
		{http.MethodTrace, "TRACE method", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/test", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
			if tt.method != http.MethodHead && tt.method != http.MethodConnect && tt.method != http.MethodOptions {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	middlewareA := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}

	_ = server.RegisterRouters(ROOT, rr, middlewareA)

	e := server.GetEcho()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test passed", rec.Body.String())
}

func TestMultiplesMiddlewares(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	middlewareA := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}

	middlewareB := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}

	_ = server.RegisterRouters(ROOT, rr, middlewareA, middlewareB)

	e := server.GetEcho()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test passed", rec.Body.String())
}

func TestMultiMiddlewaresAndMultiGroups(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	middlewareA := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}

	middlewareB := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}

	_ = server.RegisterRouters(V1, rr, middlewareA)
	_ = server.RegisterRouters(V2, rr, middlewareB)

	e := server.GetEcho()

	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/v2/test", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test passed", rec.Body.String())

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "test passed", rec2.Body.String())
}

func TestMultiMiddlewaresAndMultiGroupsIntercept(t *testing.T) {
	server, _ := NewServer()
	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	middlewareA := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("middlewareA", true)
			return next(c)
		}
	}

	middlewareB := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get("middlewareA") == true {
				return c.String(http.StatusInternalServerError, "middlewareA intercepted")
			}
			return next(c)
		}
	}

	// Register the same route with middlewareA for V1
	_ = server.RegisterRouters(V1, rr, middlewareA)

	// Register the same route with both middlewareA and middlewareB for V2
	_ = server.RegisterRouters(V2, rr, middlewareA, middlewareB)

	e := server.GetEcho()

	// Test V1
	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test passed", rec.Body.String())

	// Test V2
	req2 := httptest.NewRequest(http.MethodGet, "/v2/test", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusInternalServerError, rec2.Code)
	assert.Equal(t, "middlewareA intercepted", rec2.Body.String())
}

func TestMiddlewareSimpleAuth(t *testing.T) {
	simpleAuthMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authToken := c.Request().Header.Get("X-Auth-Token")
			if authToken != "secret-token" {
				return c.String(http.StatusUnauthorized, "Unauthorized")
			}
			return next(c)
		}
	}

	server, _ := NewServer()

	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	docs := NewRouters()
	docs.AddRouter("", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "docs content")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)
	_ = server.RegisterRouters(DOCS, docs, simpleAuthMiddleware)

	e := server.GetEcho()

	tests := []struct {
		name         string
		method       string
		path         string
		authToken    string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Unauthenticated access to /docs",
			method:       http.MethodGet,
			path:         "/docs",
			authToken:    "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: "Unauthorized",
		},
		{
			name:         "Authenticated access to /docs",
			method:       http.MethodGet,
			path:         "/docs",
			authToken:    "secret-token",
			expectedCode: http.StatusOK,
			expectedBody: "docs content",
		},
		{
			name:         "Unauthenticated access to /test",
			method:       http.MethodGet,
			path:         "/test",
			authToken:    "",
			expectedCode: http.StatusOK,
			expectedBody: "test passed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authToken != "" {
				req.Header.Set("X-Auth-Token", tt.authToken)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestInvalidEngineType(t *testing.T) {
	server, _ := NewServer()

	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	err := server.registerRouters(nil, rr)
	assert.Error(t, err)
}

func TestInvalidGroupType(t *testing.T) {
	server, _ := NewServer()

	rr := NewRouters()
	rr.AddRouter("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	err := server.RegisterRouters(999, rr)
	assert.Error(t, err)
}

func TestInvalidGroupTypeWithFixedPath(t *testing.T) {
	server, _ := NewServer()

	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	err := server.RegisterRouters(999, rr)
	assert.Error(t, err)
}

func TestRegisterRouterGetRouterFx(t *testing.T) {
	server, _ := NewServer()

	rr := NewRouters()
	rr.SetPathFixed("/api")
	rr.AddRouterFx("/test", map[string]HandlerFunc{
		http.MethodGet: func(c Context) error {
			return c.String(http.StatusOK, "test passed")
		},
	})

	_ = server.RegisterRouters(ROOT, rr)

	routes := rr.GetRoutersFx()
	assert.Len(t, routes, 1)

}

func TestSeverNewContext(t *testing.T) {
	server, _ := NewServer()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := server.NewContext(req, rec)

	assert.NotNil(t, ctx)
}

func TestServerGracefulShutdown(t *testing.T) {
	server, _ := NewServer()

	go server.Start()
	time.Sleep(1 * time.Second)

	assert.NoError(t, server.GracefulShutdown())
}

func TestNewServerReturnNil(t *testing.T) {
	server, err := NewServer()
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

func TestServerCloseReturnError(t *testing.T) {
	server, _ := NewServer()
	assert.NoError(t, server.Close())
}

func TestNewServerParamsWithNil(t *testing.T) {
	params, err := newServerParams()
	assert.NoError(t, err)
	assert.NotNil(t, params)
}
