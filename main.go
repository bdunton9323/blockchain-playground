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
		controllers.DeliverOrder(ctx, *privateKey)
	})

	// this is really just a hook for me to test
	router.POST("/contract", func(ctx *gin.Context) {
		controllers.DeployContract(ctx, *privateKey)
	})

	router.POST("/contract/execute", func(ctx *gin.Context) {
		controllers.ExecuteContract(ctx, *privateKey)
	})

	router.GET("/contract/value", func(ctx *gin.Context) {
		controllers.GetContractValue(ctx, *privateKey)
	})

	router.Run(":3000")
}
