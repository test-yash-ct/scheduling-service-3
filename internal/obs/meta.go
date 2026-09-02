package obs

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Metadata struct {
	Service   string
	Version   string
	BuildTime string
	GitSHA    string
}

func MetadataFromEnv(service string) Metadata {
	return Metadata{
		Service:   service,
		Version:   envOr("SERVICE_VERSION", "dev"),
		BuildTime: envOr("BUILD_TIME", "unknown"),
		GitSHA:    envOr("GIT_SHA", "unknown"),
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

type metaResponse struct {
	Service   string `json:"service"`
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitSHA    string `json:"git_sha"`
}

func MetaHandler(meta Metadata) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, metaResponse{
			Service:   meta.Service,
			Version:   meta.Version,
			BuildTime: meta.BuildTime,
			GitSHA:    meta.GitSHA,
		})
	}
}
