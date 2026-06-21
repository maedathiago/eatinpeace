package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/maedathiago/eatinpeace/internal/application"
	"github.com/maedathiago/eatinpeace/internal/httpapi"
	"github.com/maedathiago/eatinpeace/internal/storage/memory"
)

func TestServesOperationalConsole(t *testing.T) {
	service := application.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `<div id="root">`) {
		t.Fatalf("console html missing React root: %s", body)
	}
}

func TestServesStaticAssets(t *testing.T) {
	service := application.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	root := httptest.NewRecorder()
	rootReq := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(root, rootReq)

	asset := regexp.MustCompile(`/static/assets/[^"]+\.js`).FindString(root.Body.String())
	if asset == "" {
		t.Fatalf("console html missing built js asset: %s", root.Body.String())
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, asset, nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d want %d", rec.Code, http.StatusOK)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "javascript") {
		t.Fatalf("content-type = %q want javascript", contentType)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("built app js was empty")
	}
}
