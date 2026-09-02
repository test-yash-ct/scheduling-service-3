package obs

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"
const HeaderTenantID = "X-Tenant-ID"

type accessLogEntry struct {
	RequestID string `json:"request_id"`
	Tenant    string `json:"tenant"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Duration  int64  `json:"duration_ms"`
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(string(requestIDKey), rid)
		c.Header(HeaderRequestID, rid)
		ctx := WithRequestID(c.Request.Context(), rid)
		if tenant := c.GetHeader(HeaderTenantID); tenant != "" {
			ctx = WithTenant(ctx, tenant)
			c.Set(string(tenantKey), tenant)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func AccessLogger() gin.HandlerFunc {
	logger := log.New(os.Stdout, "", 0)
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		rid, _ := c.Get(string(requestIDKey))
		requestID, _ := rid.(string)
		tenant, _ := c.Get(string(tenantKey))
		tenantStr, _ := tenant.(string)
		entry := accessLogEntry{
			RequestID: requestID,
			Tenant:    tenantStr,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    c.Writer.Status(),
			Duration:  time.Since(start).Milliseconds(),
		}
		b, _ := json.Marshal(entry)
		logger.Println(string(b))
	}
}
