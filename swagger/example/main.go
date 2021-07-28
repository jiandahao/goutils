package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jiandahao/goutils/swagger"
)

func main() {
	eg := gin.New()
	eg.GET("/swagger/docs/*any", swagger.GinHandler(
		swagger.LoadFiles("./", "/swagger/docs"), // 加载swagger json文件
	))
	eg.Run(":8083")
}
