// Package internal 应用装配：App 由 wire 生成的 InitializeApp 组装。
package internal

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// App 应用根对象。
type App struct {
	Router *gin.Engine
	DB     *gorm.DB
}

func NewApp(r *gin.Engine, db *gorm.DB) *App {
	return &App{Router: r, DB: db}
}
