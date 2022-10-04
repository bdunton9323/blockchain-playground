package controllers

import (
	"fmt"
	"strings"

	"github.com/bdunton9323/blockchain-playground/contract"
	"github.com/bdunton9323/blockchain-playground/orders"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// The controller itself
type OrderController struct {
	NodeUrl          string
	ServerPrivateKey string
	OrderRepository  *orders.MariaDBOrderRepository
}

// The request body for updating the status of an order
type OrderUpdateRequest struct {
	Status string
}

// Metadata about the order that was placed
type CreateOrderResponse struct {
	Address         string `json:"address"`
	TokenId         string `json:"message"`
	ContractAddress string `json:"contractAddress"`
	OrderId         string `json:"orderId"`
}

// Error response from the API
type ApiError struct {
	Error string `json:"error"`
}

// PingExample godoc
// @Summary      ping example
// @Description  Places an order that can later be delivered
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        itemId        query  string  true  "The item to order"
// @Param        buyerAddress  query  string  true  "the Ethereum address of the user who can accept the delivery"
// @Success      200  {object}  CreateOrderResponse
// @Failure      400  {object}  ApiError
// @Failure      404  {object}  ApiError
// @Failure      500  {object}  ApiError
// @Router       /order [post]
func (_Controller *OrderController) CreateOrder(ctx *gin.Context) {

	if !validateArgs(ctx, "itemId", "buyerAddress") {
		return
	}

	itemId := ctx.Query("itemId")
	if len(itemId) == 0 {
		ctx.JSON(400, gin.H{
			"error": "itemId was not given",
		})
		return
	}

	// the user who is allowed to receive the shipment
	userAddress := ctx.Query("buyerAddress")

	// make up a price since there is no database of inventory
	price := int64(0)
	tokenId, address, err := contract.DeployContractAndMintNFT(
		_Controller.ServerPrivateKey,
		_Controller.NodeUrl,
		price,
		userAddress)

	if err != nil {
		errorResponse(ctx, 500, err.Error())
		return
	}

	order := &orders.Order{
		OrderId:      uuid.New().String(),
		ItemId:       itemId,
		ItemName:     "socks",
		Price:        price,
		TokenAddress: *address,
		TokenId:      tokenId.Int64(),
		Delivered:    false,
	}

	err = _Controller.OrderRepository.CreateOrder(order)
	if err != nil {
		errorResponse(ctx, 500, "Error writing order to database")
		return
	}

	ctx.JSON(200, CreateOrderResponse{
		Address:         *address,
		TokenId:         tokenId.String(),
		ContractAddress: *address,
		OrderId:         order.OrderId,
	})
}

func (_Controller *OrderController) DeliverOrder(ctx *gin.Context) {
	if !validateArgs(ctx, "customerKey") {
		return
	}

	// The customer is signing for the order, and they have a different key than
	// the one loaded into the server. I recognize that transmitting the private
	// key to a server is a terrible idea, but given the lack of time, it at least
	// demonstrates the smart contract's functionality.
	customerPrivateKey := ctx.Query("customerKey")

	orderId := ctx.Param("orderId")
	log.Infof("Delivering order [%v]", orderId)

	order, err := _Controller.OrderRepository.GetOrder(orderId)
	if err != nil {
		errorResponse(ctx, 500, err.Error())
		return
	} else if order == nil || order.Delivered {
		orderNotFoundResponse(ctx, orderId)
		return
	}

	// buy the token from the vendor, thereby accepting delivery of the package
	err = contract.BuyNFT(order.TokenAddress, order.TokenId, customerPrivateKey, _Controller.NodeUrl)
	if err != nil {
		errorResponse(ctx, 500, err.Error())
		return
	}

	_Controller.OrderRepository.MarkOrderDelivered(orderId)

	ctx.JSON(200, gin.H{
		"status": "delivered",
	})

	return
}

// Determines who currently owns the deliver token - the vendor or the customer.
// Returns the owner's address.
func (_Controller *OrderController) GetDeliveryTokenOwner(ctx *gin.Context) {

	orderId := ctx.Param("orderId")
	order, _ := _Controller.OrderRepository.GetOrder(orderId)
	if order == nil {
		orderNotFoundResponse(ctx, orderId)
		return
	}

	owner, err := contract.GetOwner(order.TokenAddress, _Controller.ServerPrivateKey)

	if err != nil {
		errorResponse(ctx, 500, err.Error())
	} else {
		ctx.JSON(200, gin.H{
			"owner": owner,
		})
	}
}

func (_Controller *OrderController) CancelOrder(ctx *gin.Context) {
	orderId := ctx.Param("orderId")
	order, _ := _Controller.OrderRepository.GetOrder(orderId)
	if order == nil {
		orderNotFoundResponse(ctx, orderId)
		return
	}

	err := contract.BurnContract(order.TokenAddress, _Controller.ServerPrivateKey)
	if err != nil {
		errorResponse(ctx, 500, "Failed to cancel the order")
	}
}

func validateArgs(ctx *gin.Context, args ...string) bool {
	var sb strings.Builder

	valid := true
	first := true
	for _, arg := range args {
		if len(ctx.Query(arg)) == 0 {
			valid = false
			if !first {
				sb.WriteString(",")
			}
			sb.WriteString(arg)
			first = false
		}
	}

	if !valid {
		errorResponse(ctx, 400, sb.String())
		return false
	}
	return true
}

func errorResponse(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, gin.H{
		"error": message,
	})
}

func orderNotFoundResponse(ctx *gin.Context, orderId string) {
	errorResponse(ctx, 404, fmt.Sprintf("Order ID [%s] does not exist", orderId))
}
