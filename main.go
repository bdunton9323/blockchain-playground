package main

import (
	"flag"

	"github.com/bdunton9323/blockchain-playground/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	privateKey := flag.String("privatekey", "", "The private key of the vendor")
	flag.Parse()

	router := gin.Default()

	router.POST("/order", func(ctx *gin.Context) {
		controllers.CreateOrder(ctx, *privateKey)
	})

	// TODO: to make this RESTful, should take the action in the post body not the URL path
	router.POST("/order/:orderId/deliver", func(ctx *gin.Context) {
		controllers.DeliverOrder(ctx)
	})

	router.GET("/order/:orderId", func(ctx *gin.Context) {
		controllers.GetOrderStatus(ctx, *privateKey)
	})

	router.Run(":3000")
}
