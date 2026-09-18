package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lyrics-server/internal/model"
	"lyrics-server/internal/req"
	"lyrics-server/internal/resp"
	"lyrics-server/internal/service"
)

type LyricHandler struct {
	svc service.LyricService
}

func NewLyricHandler(s service.LyricService) *LyricHandler {
	return &LyricHandler{svc: s}
}

func (h *LyricHandler) Parse(c *gin.Context) {
	var r req.ParseReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	result, err := h.svc.Parse(c.Request.Context(), r.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(resp.ParseResp{
		Language: string(result.Language),
		Lines:    result.Lines,
	}))
}

func (h *LyricHandler) Translate(c *gin.Context) {
	var r req.TranslateReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	translations, err := h.svc.Translate(c.Request.Context(), r.Source, r.Target, r.Lines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(resp.TranslateResp{Translations: translations}))
}

func (h *LyricHandler) Tokenize(c *gin.Context) {
	var r req.TokenizeReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	tokens, err := h.svc.Tokenize(c.Request.Context(), model.Language(r.Language), r.Line)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(resp.TokenizeResp{Tokens: tokens}))
}
