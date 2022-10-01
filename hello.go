package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(200, "Hello World!")
	})

	api := router.Group("/api")
	api.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.Run(":3000")
}
