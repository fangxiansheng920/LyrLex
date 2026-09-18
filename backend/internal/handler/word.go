package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-server/internal/req"
	"lyrics-server/internal/resp"
	"lyrics-server/internal/service"
)

type WordHandler struct {
	svc service.WordService
}

func NewWordHandler(s service.WordService) *WordHandler {
	return &WordHandler{svc: s}
}

func (h *WordHandler) Lookup(c *gin.Context) {
	var r req.LookupReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	result, err := h.svc.Lookup(c.Request.Context(), r.Language, r.Word, r.Context, r.Force)
	if err != nil {
		c.JSON(http.StatusNotFound, resp.Error(http.StatusNotFound, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(result))
}

func (h *WordHandler) Audio(c *gin.Context) {
	language := c.Param("language")
	word := c.Param("word")
	data, err := h.svc.Audio(c.Request.Context(), language, word)
	if err != nil {
		c.JSON(http.StatusNotFound, resp.Error(http.StatusNotFound, err.Error()))
		return
	}
	c.Data(http.StatusOK, "audio/mpeg", data)
}
