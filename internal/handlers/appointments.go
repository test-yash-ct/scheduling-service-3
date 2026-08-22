package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthops/scheduling-service/internal/audit"
	"github.com/healthops/scheduling-service/internal/middleware"
	"github.com/healthops/scheduling-service/internal/store"
)

type AppointmentAPI struct {
	Store *store.AppointmentStore
	Audit *audit.Logger
}

func (a *AppointmentAPI) Register(r *gin.RouterGroup) {
	r.POST("/appointments", a.Create)
	r.POST("/appointments/:id/cancel", a.Cancel)
}

type createBody struct {
	ProviderID string    `json:"provider_id"`
	PatientID  string    `json:"patient_id"`
	SlotStart  time.Time `json:"slot_start"`
}

func (a *AppointmentAPI) Create(c *gin.Context) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if !claims.HasRole("appointment:book") {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	idem, ok := middleware.IdempotencyKey(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idempotency_key_required"})
		return
	}
	var body createBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	if !middleware.ValidID(body.ProviderID) || !middleware.ValidID(body.PatientID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}
	if body.SlotStart.IsZero() || body.SlotStart.Before(time.Now().Add(-time.Minute)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_slot"})
		return
	}
	if claims.Sub != body.PatientID && !claims.HasRole("appointment:admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	id := uuid.NewString()
	err := a.Store.BookSlot(c.Request.Context(), id, claims.Tenant, body.ProviderID, body.PatientID, body.SlotStart, idem)
	if errors.Is(err, store.ErrSlotTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": "slot_taken"})
		return
	}
	if errors.Is(err, store.ErrIdempotency) {
		c.JSON(http.StatusConflict, gin.H{"error": "idempotency_conflict"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist_failed"})
		return
	}
	a.emit(c, claims.Sub, claims.Tenant, id, "appointment.create", "ok")
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *AppointmentAPI) Cancel(c *gin.Context) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id := c.Param("id")
	if !middleware.ValidID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}
	appt, err := a.Store.GetByID(c.Request.Context(), claims.Tenant, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	if appt.PatientID != claims.Sub && !claims.HasRole("appointment:cancel") && !claims.HasRole("appointment:admin") {
		a.emit(c, claims.Sub, claims.Tenant, id, "appointment.cancel", "denied")
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if err := a.Store.Cancel(c.Request.Context(), claims.Tenant, id); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "invalid_state"})
		return
	}
	a.emit(c, claims.Sub, claims.Tenant, id, "appointment.cancel", "ok")
	c.Status(http.StatusNoContent)
}

func (a *AppointmentAPI) emit(c *gin.Context, actor, tenant, object, action, outcome string) {
	rid, _ := c.Get("request_id")
	ridStr, _ := rid.(string)
	if a.Audit != nil {
		a.Audit.Emit(audit.Event{Actor: actor, Tenant: tenant, Action: action, ObjectID: object, Outcome: outcome, RequestID: ridStr})
	}
}
