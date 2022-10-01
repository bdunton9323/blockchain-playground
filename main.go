package main

import (
	"github.com/gin-gonic/gin"
	"github.com/bdunton9323/blockchain-playground/controllers"
)

func main() {
	router := gin.Default()

	router.GET("/hello", func(ctx *gin.Context) {
		controllers.HelloController(ctx)
	})

	api := router.Group("/api")
	api.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.Run(":3000")
}
