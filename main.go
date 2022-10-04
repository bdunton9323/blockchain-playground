package main

import (
	"flag"

	"github.com/bdunton9323/blockchain-playground/controllers"
	"github.com/bdunton9323/blockchain-playground/orders"
	"github.com/gin-gonic/gin"
)

var ethNodeUrl = "http://172.13.3.1:8545"
var dbHost = "127.0.0.1:3306"
var dbName = "orderdb"
var dbUser = "db_user"
var dbPassword = "mysqlPassword"

func main() {
	privateKey := flag.String("privatekey", "", "The private key of the vendor")
	flag.Parse()

	orderRepo, err := orders.NewMariaDBOrderRepository(dbHost, dbName, dbUser, dbPassword)
	if err != nil {
		panic(err.Error())
	}

	var orderController *controllers.OrderController = new(controllers.OrderController)
	orderController.ServerPrivateKey = privateKey
	orderController.NodeUrl = &ethNodeUrl
	orderController.OrderRepository = orderRepo

	router := gin.Default()
	router.POST("/order", func(ctx *gin.Context) {
		orderController.CreateOrder(ctx)
	})

	// TODO: make this RESTful by taking the 'delivered' state in the post body, not the URL path
	router.POST("/order/:orderId/deliver", func(ctx *gin.Context) {
		orderController.DeliverOrder(ctx)
	})

	router.GET("/order/:orderId/owner", func(ctx *gin.Context) {
		orderController.GetDeliveryTokenOwner(ctx)
	})

	// TODO: make this RESTful by taking the 'canceled' state in the POST body.
	router.POST("/order/:orderId/cancel", func(ctx *gin.Context) {
		orderController.CancelOrder(ctx)
	})

	router.Run(":3000")
}
