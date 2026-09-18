// Package middleware 日志、错误处理、recovery。
package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"lyrics-server/internal/resp"
)

// Logger 简单访问日志。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		log.Printf("%s %s %d %s", c.Request.Method, path, c.Writer.Status(), time.Since(start))
	}
}

// Recovery 捕获 panic，返回统一错误结构。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					resp.Error(http.StatusInternalServerError, "internal server error"))
			}
		}()
		c.Next()
	}
}

// CORS 允许 Electron 渲染层（file:// 源）访问 localhost 后端。
// 后端仅绑定 127.0.0.1，放开跨域是安全的。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
