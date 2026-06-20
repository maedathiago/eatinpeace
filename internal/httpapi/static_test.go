package httpapi_test

import (
	"net/http"
	"net/http/httptest"
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
	if !strings.Contains(body, "Console operacional P0") {
		t.Fatalf("console html missing title: %s", body)
	}
}

func TestServesStaticAssets(t *testing.T) {
	service := application.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "openSession") {
		t.Fatal("app js was not served")
	}
}
