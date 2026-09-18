package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-server/internal/resp"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check 健康检查。
func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, resp.OK(gin.H{"status": "ok"}))
}
