package controllers

import (
	_ "github.com/bdunton9323/blockchain-playground/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

// @title           Vendor API
// @version         1.0
// @description     These APIs allow the client to order items from the vendor
// @license.name    MIT
// @license.url     https://github.com/bdunton9323/blockchain-playground/blob/main/LICENSE
// @host            localhost:8080
// @BasePath        /api/v1
type ApiRouter struct {
	orderController *OrderController
}

// Constructs a new API router that dispatches to the given controller
func NewApiRouter(orderController *OrderController) *ApiRouter {
	return &ApiRouter{
		orderController: orderController,
	}
}

// Starts listening for HTTP requests
func (_apiRouter *ApiRouter) Start() {
	router := gin.Default()

	router.POST("/api/v1/order", func(ctx *gin.Context) {
		_apiRouter.orderController.CreateOrder(ctx)
	})

	router.POST("/api/v1/order/:orderId", func(ctx *gin.Context) {
		_apiRouter.orderController.UpdateOrderStatus(ctx)
	})

	router.POST("/api/v1/payment/order/:orderId", func(ctx *gin.Context) {
		_apiRouter.orderController.PayForOrder(ctx)
	})

	router.GET("/api/v1/order/:orderId/owner", func(ctx *gin.Context) {
		_apiRouter.orderController.GetDeliveryTokenOwner(ctx)
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run(":8080")
}
