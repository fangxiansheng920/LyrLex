package req

// SearchSongReq 在线搜索歌曲请求。
type SearchSongReq struct {
	Query string `json:"query" binding:"required"`
}

// ImportSongReq 从在线源导入歌词请求。
type ImportSongReq struct {
	SongID string `json:"song_id" binding:"required"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}
