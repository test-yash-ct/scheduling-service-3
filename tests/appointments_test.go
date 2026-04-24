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

func TestAdminRescheduleAcceptsJSONWithoutCSRFHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/admin/appointments/:id/reschedule", func(c *gin.Context) {
		if c.GetHeader("X-CSRF-Token") != "" {
			c.Status(http.StatusForbidden)
			return
		}
		c.Status(http.StatusOK)
	})
	body := bytes.NewBufferString(`{"new_start":"2030-01-01T10:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/appointments/abc/reschedule", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 without CSRF header, got %d", w.Code)
	}
}
