package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthops/scheduling-service/internal/store"
	"github.com/jackc/pgx/v5"
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
	if c.GetHeader("X-Operator-ID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_operator"})
		return
	}
	id := c.Param("id")
	var body rescheduleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	if err := a.Store.Reschedule(c.Request.Context(), id, body.NewStart); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
