package main

import (
	"flag"

	"github.com/bdunton9323/blockchain-playground/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	privateKey := flag.String("privatekey", "", "The user's private key")
	flag.Parse()

	router := gin.Default()

	// this is really just a hook for me to test
	router.POST("/contract", func(ctx *gin.Context) {
		controllers.DeployContract(ctx, *privateKey)
	})

	// some examples for my own reference

	// router.GET("/hello", func(ctx *gin.Context) {
	// 	controllers.HelloController(ctx)
	// })

	// api := router.Group("/api")
	// api.GET("/ping", func(ctx *gin.Context) {
	// 	ctx.JSON(200, gin.H{
	// 		"message": "pong",
	// 	})
	// })

	router.Run(":3000")
}
