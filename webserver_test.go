package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFuelEndpointsWithoutSharedMemory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ws := NewWebserver(":0")
	ws.RegisterSharedMemoryEndpoint(&SharedMemReader{}) // never opened

	for _, path := range []string{"/fuel/flow", "/fuel/state"} {
		rec := httptest.NewRecorder()
		ws.r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("%s: got status %d, want %d", path, rec.Code, http.StatusServiceUnavailable)
		}
	}
}
