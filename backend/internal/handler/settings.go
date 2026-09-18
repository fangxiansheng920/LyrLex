package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-server/internal/req"
	"lyrics-server/internal/resp"
	"lyrics-server/internal/service"
)

type SettingsHandler struct {
	svc service.SettingsService
}

func NewSettingsHandler(s service.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: s}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	settings, err := h.svc.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(settings))
}

func (h *SettingsHandler) Save(c *gin.Context) {
	var r req.UpdateSettingsReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	// 先读取再局部更新，避免覆盖未传字段
	current, err := h.svc.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	if r.APIKeys != nil {
		current.APIKeys = service.APIConfig{
			YoudaoAppKey:    r.APIKeys.YoudaoAppKey,
			YoudaoAppSecret: r.APIKeys.YoudaoAppSecret,
			BaiduAppID:      r.APIKeys.BaiduAppID,
			BaiduSecret:     r.APIKeys.BaiduSecret,
			DeepLKey:        r.APIKeys.DeepLKey,
		}
	}
	if r.PronunciationSource != "" {
		current.PronunciationSource = r.PronunciationSource
	}
	if err := h.svc.Save(c.Request.Context(), current); err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(current))
}

func (h *SettingsHandler) DataDir(c *gin.Context) {
	c.JSON(http.StatusOK, resp.OK(gin.H{"path": h.svc.DataDir()}))
}

// TestTranslation 测试当前翻译源（有道）是否可用。
func (h *SettingsHandler) TestTranslation(c *gin.Context) {
	c.JSON(http.StatusOK, resp.OK(h.svc.TestTranslation(c.Request.Context())))
}
