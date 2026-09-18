package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lyrics-server/internal/model"
	"lyrics-server/internal/provider"
	"lyrics-server/internal/req"
	"lyrics-server/internal/resp"
	"lyrics-server/internal/service"
)

type SongHandler struct {
	svc service.LyricService
}

func NewSongHandler(s service.LyricService) *SongHandler {
	return &SongHandler{svc: s}
}

func (h *SongHandler) Save(c *gin.Context) {
	var r req.SaveSongReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	lines := make([]service.SongLineIn, 0, len(r.Lines))
	for _, l := range r.Lines {
		tokens := make([]provider.Token, 0, len(l.Tokens))
		for _, t := range l.Tokens {
			tokens = append(tokens, provider.Token{
				Surface: t.Surface,
				Lemma:   t.Lemma,
				POS:     t.POS,
				Reading: t.Reading,
				Romaji:  t.Romaji,
				Accent:  t.Accent,
				GroupID: t.GroupID,
				Addable: t.Addable,
			})
		}
		lines = append(lines, service.SongLineIn{
			Text:          l.Text,
			TranslationZH: l.TranslationZH,
			Tokens:        tokens,
		})
	}
	song, err := h.svc.SaveSong(c.Request.Context(), r.Title, r.Artist, model.Language(r.Language), lines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(song))
}

func (h *SongHandler) List(c *gin.Context) {
	query := c.Query("q")
	songs, err := h.svc.ListSongs(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(songs))
}

func (h *SongHandler) Search(c *gin.Context) {
	var r req.SearchSongReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	metas, err := h.svc.SearchSongs(c.Request.Context(), r.Query)
	if err != nil {
		c.JSON(http.StatusBadGateway, resp.Error(http.StatusBadGateway, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(metas))
}

func (h *SongHandler) Import(c *gin.Context) {
	var r req.ImportSongReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	song, err := h.svc.ImportSong(c.Request.Context(), provider.SongMeta{
		ID:     r.SongID,
		Title:  r.Title,
		Artist: r.Artist,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, resp.Error(http.StatusBadGateway, err.Error()))
		return
	}
	lines, err := h.svc.GetSongLines(c.Request.Context(), song.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(gin.H{"song": song, "lines": lines}))
}

func (h *SongHandler) GetLyrics(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: id"))
		return
	}
	song, err := h.svc.GetSong(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, resp.Error(http.StatusNotFound, err.Error()))
		return
	}
	lines, err := h.svc.GetSongLines(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(gin.H{"song": song, "lines": lines}))
}

// Delete 删除本地歌曲及其歌词。
func (h *SongHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: id"))
		return
	}
	if err := h.svc.DeleteSong(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(nil))
}

// BatchDelete 批量删除本地歌曲。
func (h *SongHandler) BatchDelete(c *gin.Context) {
	var r req.BatchDeleteSongsReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}
	if err := h.svc.DeleteSongs(c.Request.Context(), r.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp.OK(nil))
}
