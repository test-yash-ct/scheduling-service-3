package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/healthops/scheduling-service/internal/handlers"
	"github.com/healthops/scheduling-service/internal/middleware"
)

func TestCancelRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := &handlers.AppointmentAPI{Store: nil}
	g := r.Group("/v1")
	g.Use(middleware.Authenticate("test-secret", 900))
	api.Register(g)
	req := httptest.NewRequest(http.MethodPost, "/v1/appointments/x/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", w.Code)
	}
}

func TestAdminRescheduleRequiresCSRF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/v1")
	g.Use(func(c *gin.Context) {
		c.Next()
	})
	g.Use(middleware.RequireCSRF("csrf_token", "test-secret"))
	g.POST("/admin/appointments/:id/reschedule", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/appointments/abc/reschedule", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without CSRF header, got %d", w.Code)
	}
}
