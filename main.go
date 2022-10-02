package main

import (
	"flag"

	"github.com/bdunton9323/blockchain-playground/controllers"
	"github.com/gin-gonic/gin"
)

var nodeUrl = "http://172.13.3.1:8545"

func main() {
	privateKey := flag.String("privatekey", "", "The private key of the vendor")
	flag.Parse()

	router := gin.Default()

	var orderController *controllers.OrderController = new(controllers.OrderController)
	orderController.ServerPrivateKey = privateKey
	orderController.NodeUrl = &nodeUrl

	router.POST("/order", func(ctx *gin.Context) {
		orderController.CreateOrder(ctx)
	})

	// TODO: to make this RESTful, should take the action in the post body not the URL path
	router.POST("/order/:orderId/deliver", func(ctx *gin.Context) {
		orderController.DeliverOrder(ctx)
	})

	router.GET("/order/:orderId", func(ctx *gin.Context) {
		orderController.GetOrderStatus(ctx)
	})

	router.Run(":3000")
}
