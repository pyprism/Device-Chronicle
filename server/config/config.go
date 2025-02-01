package config

import (
	"device-chronicle-server/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	limits "github.com/gin-contrib/size"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

var Logger *zap.Logger

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*.html")

	Logger, _ = zap.NewProduction()
	router.Use(ginzap.Ginzap(Logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(Logger, true))
	if utils.GetEnv("DEBUG", "false") == "false" {
		gin.SetMode(gin.ReleaseMode)
	}

	router.SetTrustedProxies([]string{"127.0.0.1"})
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1"}
	router.Use(gin.Recovery())
	router.Use(cors.New(config))
	router.Use(gzip.Gzip(gzip.BestCompression))
	router.Use(limits.RequestSizeLimiter(10000)) // 10KB

	RegisterRoutes(router)

	return router
}

func Init() {
	r := NewRouter()
	serverPort := utils.GetEnv("SERVER_PORT", "8000")
	Logger.Info("Server running on http://127.0.0.1:" + serverPort)
	err := r.Run(":" + serverPort)
	if err != nil {
		panic(err)
	}
}
