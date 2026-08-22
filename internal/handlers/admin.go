package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthops/scheduling-service/internal/audit"
	"github.com/healthops/scheduling-service/internal/middleware"
	"github.com/healthops/scheduling-service/internal/store"
)

type AdminAPI struct {
	Store *store.AppointmentStore
	Audit *audit.Logger
}

func (a *AdminAPI) Register(r *gin.RouterGroup) {
	r.POST("/admin/appointments/:id/reschedule", a.Reschedule)
}

type rescheduleBody struct {
	NewStart time.Time `json:"new_start"`
}

func (a *AdminAPI) Reschedule(c *gin.Context) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok || !claims.HasRole("appointment:admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	id := c.Param("id")
	if !middleware.ValidID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}
	var body rescheduleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	if body.NewStart.IsZero() || body.NewStart.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_slot"})
		return
	}
	err := a.Store.Reschedule(c.Request.Context(), claims.Tenant, id, body.NewStart)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	if errors.Is(err, store.ErrSlotTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": "slot_taken"})
		return
	}
	if errors.Is(err, store.ErrInvalidState) {
		c.JSON(http.StatusConflict, gin.H{"error": "invalid_state"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed"})
		return
	}
	rid, _ := c.Get("request_id")
	ridStr, _ := rid.(string)
	if a.Audit != nil {
		a.Audit.Emit(audit.Event{Actor: claims.Sub, Tenant: claims.Tenant, Action: "appointment.reschedule", ObjectID: id, Outcome: "ok", RequestID: ridStr})
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
