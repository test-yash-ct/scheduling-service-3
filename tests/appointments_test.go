package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/healthops/scheduling-service/internal/handlers"
	"github.com/healthops/scheduling-service/internal/obs"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(obs.RequestID())
	r.Use(obs.AccessLogger())
	r.GET("/meta", obs.MetaHandler(obs.Metadata{
		Service:   "scheduling-service",
		Version:   "test",
		BuildTime: "2026-01-01T00:00:00Z",
		GitSHA:    "abc123",
	}))
	return r
}

func TestMetaEndpoint(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/meta", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["service"] != "scheduling-service" {
		t.Fatalf("service %q", body["service"])
	}
}

func TestRequestIDEcho(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/meta", nil)
	req.Header.Set(obs.HeaderRequestID, "sched-req-99")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get(obs.HeaderRequestID); got != "sched-req-99" {
		t.Fatalf("expected echoed request id, got %q", got)
	}
}

func TestCancelRequiresOperatorHeader(t *testing.T) {
	r := setupRouter()
	api := &handlers.AppointmentAPI{Store: nil}
	g := r.Group("/v1")
	api.Register(g)
	req := httptest.NewRequest(http.MethodPost, "/v1/appointments/x/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}

func TestAdminRescheduleRequiresOperatorHeader(t *testing.T) {
	r := setupRouter()
	admin := &handlers.AdminAPI{Store: nil}
	g := r.Group("/v1")
	admin.Register(g)
	body := bytes.NewBufferString(`{"new_start":"2030-01-01T10:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/appointments/abc/reschedule", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without operator header, got %d", w.Code)
	}
}
