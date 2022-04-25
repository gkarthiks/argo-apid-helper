package config

import "github.com/gin-gonic/gin"

func InitializeRouter() {
	Router = gin.New()
	Router.Use(Logger(Log), gin.Recovery())
}
