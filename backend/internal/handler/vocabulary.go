package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lyrics-server/internal/req"
	"lyrics-server/internal/resp"
	"lyrics-server/internal/service"
)

type VocabularyHandler struct {
	svc service.VocabularyService
}

func NewVocabularyHandler(s service.VocabularyService) *VocabularyHandler {
	return &VocabularyHandler{svc: s}
}

func (h *VocabularyHandler) Add(c *gin.Context) {
	var r req.AddVocabularyReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	item, err := h.svc.Add(c.Request.Context(), toServiceAddReq(&r))
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(item))
}

func (h *VocabularyHandler) List(c *gin.Context) {
	language := c.DefaultQuery("language", "en")
	query := c.Query("q")
	songID, _ := strconv.ParseInt(c.Query("song_id"), 10, 64)
	items, err := h.svc.List(c.Request.Context(), language, query, songID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(items))
}

func (h *VocabularyHandler) Get(c *gin.Context) {
	language := c.Param("language")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: id"))
		return
	}
	detail, err := h.svc.Get(c.Request.Context(), language, id)
	if err != nil {
		c.JSON(http.StatusNotFound, resp.Error(http.StatusNotFound, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(detail))
}

func (h *VocabularyHandler) Delete(c *gin.Context) {
	language := c.Param("language")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), language, id); err != nil {
		c.JSON(http.StatusNotFound, resp.Error(http.StatusNotFound, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(nil))
}

func toServiceAddReq(r *req.AddVocabularyReq) *service.AddVocabularyReq {
	out := &service.AddVocabularyReq{
		Language: r.Language,
		Word:     r.Word,
		Lemma:    r.Lemma,
		Kana:     r.Kana,
		Romaji:   r.Romaji,
	}
	for _, m := range r.Meanings {
		out.Meanings = append(out.Meanings, service.MeaningIn{POS: m.POS, Meaning: m.Meaning})
	}
	if r.Context != nil {
		out.Context = &service.ContextIn{
			Text:          r.Context.Text,
			TranslationZH: r.Context.TranslationZH,
			SongID:        r.Context.SongID,
			LineIndex:     r.Context.LineIndex,
		}
	}
	return out
}
