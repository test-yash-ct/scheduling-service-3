package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/healthops/scheduling-service/internal/handlers"
)

func TestCancelRequiresOperatorHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
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
	gin.SetMode(gin.TestMode)
	r := gin.New()
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
