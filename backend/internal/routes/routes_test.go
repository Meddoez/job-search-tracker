package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesRegistersPublicAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router, nil, "", nil, nil)

	want := map[string]bool{
		http.MethodGet + " /ping":                       true,
		http.MethodPost + " /jobs/search":               true,
		http.MethodGet + " /profiles":                   true,
		http.MethodPost + " /profiles":                  true,
		http.MethodGet + " /profiles/:id":               true,
		http.MethodPost + " /profiles/:id/jobs":         true,
		http.MethodPatch + " /profiles/:id/jobs/:jobId": true,
		http.MethodDelete + " /profiles/:id":            true,
		http.MethodGet + " /profiles/:id/jobs":          true,
	}
	got := make(map[string]bool)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected routes: %#v", got)
	}
	for route := range want {
		if !got[route] {
			t.Errorf("missing route %s", route)
		}
	}
}

func TestPingReturnsPong(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router, nil, "", nil, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if response.Body.String() != `{"message":"pong"}` {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}
