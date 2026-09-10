package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthops/scheduling-service/internal/obs"
	"github.com/healthops/scheduling-service/internal/service"
	"github.com/healthops/scheduling-service/internal/store"
)

type AppointmentAPI struct {
	Store   *store.AppointmentStore
	Booking *service.Booking
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
	ctx := c.Request.Context()
	requestID := obs.RequestIDFromContext(ctx)

	var body createBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	if body.TenantID != "" {
		ctx = obs.WithTenant(ctx, body.TenantID)
		c.Request = c.Request.WithContext(ctx)
	}
	id, err := a.Booking.Book(ctx, body.TenantID, body.ProviderID, body.PatientID, body.SlotStart)
	if err != nil {
		if errors.Is(err, store.ErrSlotTaken) {
			logStoreEvent(requestID, body.TenantID, "slot_taken")
			c.JSON(http.StatusConflict, gin.H{"error": "slot_taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist_failed"})
		return
	}
	logStoreEvent(requestID, body.TenantID, "appointment_created")
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *AppointmentAPI) Cancel(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := obs.RequestIDFromContext(ctx)
	if c.GetHeader("X-Operator-ID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_operator"})
		return
	}
	id := c.Param("id")
	if err := a.Store.Cancel(ctx, id); err != nil {
		logStoreEvent(requestID, obs.TenantFromContext(ctx), "cancel_not_found")
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	logStoreEvent(requestID, obs.TenantFromContext(ctx), "appointment_cancelled")
	c.Status(http.StatusNoContent)
}

func logStoreEvent(requestID, tenant, event string) {
	entry := map[string]string{
		"event":      event,
		"request_id": requestID,
		"tenant":     tenant,
	}
	b, _ := json.Marshal(entry)
	log.New(os.Stdout, "", 0).Println(string(b))
}
