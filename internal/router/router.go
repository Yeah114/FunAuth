package router

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/Yeah114/FunAuth/internal/handlers"
)

func NewRouter() *gin.Engine {
	// 确保 gin 的默认日志写到 stdout
	gin.DefaultWriter = io.MultiWriter(os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(os.Stdout)
	log.SetOutput(os.Stdout)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	_ = r.SetTrustedProxies([]string{"127.0.0.1"})

	api := r.Group("/api")
	handlers.RegisterNewRoutes(api)
	handlers.RegisterPhoenixRoutes(api)

	return r
}
