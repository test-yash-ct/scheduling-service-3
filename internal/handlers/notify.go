package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/healthops/scheduling-service/internal/middleware"
	"github.com/healthops/scheduling-service/internal/notify"
)

type NotifyAPI struct {
	Allowlist []string
}

func (n *NotifyAPI) Register(r *gin.RouterGroup) {
	r.POST("/notify/test", n.Test)
}

type notifyBody struct {
	URL string `json:"url"`
}

func (n *NotifyAPI) Test(c *gin.Context) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok || !claims.HasRole("appointment:admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body notifyBody
	if err := c.ShouldBindJSON(&body); err != nil || body.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	b, code, err := notify.ProbeCallback(body.URL, n.Allowlist)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "fetch_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": code, "bytes": len(b)})
}
