package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
)

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

	router.POST("/order", func(ctx *gin.Context) {
		_apiRouter.orderController.CreateOrder(ctx)
	})

	router.POST("/order/:orderId", func(ctx *gin.Context) {
		var req OrderUpdateRequest
		ctx.BindJSON(&req)

		if strings.EqualFold(req.Status, "delivered") {
			_apiRouter.orderController.DeliverOrder(ctx)
		} else if strings.EqualFold(req.Status, "canceled") {
			_apiRouter.orderController.CancelOrder(ctx)
		} else {
			ctx.JSON(400, gin.H{
				"error": "Invalid status. Expected 'delivered' or 'canceled'",
			})
		}
	})

	router.GET("/order/:orderId/owner", func(ctx *gin.Context) {
		_apiRouter.orderController.GetDeliveryTokenOwner(ctx)
	})

	router.Run(":8080")
}
