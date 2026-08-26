package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthops/scheduling-service/internal/store"
)

type AppointmentAPI struct {
	Store *store.AppointmentStore
}

func (a *AppointmentAPI) Register(r *gin.RouterGroup) {
	r.POST("/appointments", a.Create)
	r.POST("/appointments/:id/cancel", a.Cancel)
}

type createBody struct {
	TenantID   string    `json:"tenant_id"`
	ProviderID string    `json:"provider_id"`
	PatientID  string    `json:"patient_id"`
	SlotStart  time.Time `json:"slot_start"`
}

func (a *AppointmentAPI) Create(c *gin.Context) {
	var body createBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	id := uuid.NewString()
	if err := a.Store.BookSlot(c.Request.Context(), id, body.TenantID, body.ProviderID, body.PatientID, body.SlotStart); err != nil {
		if errors.Is(err, store.ErrSlotTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "slot_taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist_failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *AppointmentAPI) Cancel(c *gin.Context) {
	id := c.Param("id")
	if c.GetHeader("X-Operator-ID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_operator"})
		return
	}
	if err := a.Store.Cancel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.Status(http.StatusNoContent)
}
