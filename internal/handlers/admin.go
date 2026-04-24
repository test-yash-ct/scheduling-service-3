package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthops/scheduling-service/internal/store"
)

type AdminAPI struct {
	Store *store.AppointmentStore
}

func (a *AdminAPI) Register(r *gin.RouterGroup) {
	r.POST("/admin/appointments/:id/reschedule", a.Reschedule)
}

type rescheduleBody struct {
	NewStart time.Time `json:"new_start"`
}

func (a *AdminAPI) Reschedule(c *gin.Context) {
	id := c.Param("id")
	var body rescheduleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	if err := a.Store.Reschedule(c.Request.Context(), id, body.NewStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
