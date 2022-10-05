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
	// the URL of the ethereum node to connect to
	NodeUrl string
	// the private key of the server's ethereum address
	ServerPrivateKey string
	// the persistence layer for the orders
	OrderRepository *orders.MariaDBOrderRepository
	// executes operations on the smart delivery contract
	ContractExecutor *contract.DeliveryContractExecutor
}

// The request body for updating the status of an order. One of "delivered" or "canceled".
type OrderUpdateRequest struct {
	Status string `json:"status"`
}

// Indicates the status of an order
type OrderStatusResponse struct {
	Status string `json:"status"`
}

// Metadata about the order that was placed
type CreateOrderResponse struct {
	Address         string `json:"address"`
	TokenId         string `json:"message"`
	ContractAddress string `json:"contractAddress"`
	OrderId         string `json:"orderId"`
}

// Indicates the address of the owner of the delivery token
type TokenOwnerResponse struct {
	Owner string `json:"owner" format:"address"`
}

// Error response from the API
type ApiError struct {
	Error string `json:"error"`
}

// CreateOrder godoc
// @Summary      Create order
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

	// the customer who is allowed to receive the shipment
	userAddress := ctx.Query("buyerAddress")

	// make up a price since there is no database of inventory
	price := int64(500)
	tokenId, address, err := _Controller.ContractExecutor.DeployContractAndMintNFT(
		_Controller.ServerPrivateKey,
		_Controller.NodeUrl,
		price,
		userAddress)

	if err != nil {
		ctx.JSON(500, ApiError{
			Error: err.Error(),
		})
		return
	}

	order := &orders.Order{
		OrderId:      uuid.New().String(),
		ItemId:       itemId,
		ItemName:     "socks",
		Price:        price,
		TokenAddress: address,
		TokenId:      tokenId.Int64(),
		Delivered:    false,
	}

	err = _Controller.OrderRepository.CreateOrder(order)
	if err != nil {
		ctx.JSON(500, ApiError{
			Error: "Error writing order to database",
		})
	} else {
		ctx.JSON(200, CreateOrderResponse{
			Address:         address,
			TokenId:         tokenId.String(),
			ContractAddress: address,
			OrderId:         order.OrderId,
		})
	}
}

// DeliverOrder  godoc
// @Summary      Update order status
// @Description  This action changes the status of an order, either by accepting delivery or canceling it
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        request        body   OrderUpdateRequest true  "Indicates the status of the order. One of ('canceled', 'delivered')"
// @Param        customerKey    query  string             false "If this is a delivery, the delivery recipient's private key (not a good idea in real life!)"
// @Param        orderId        path   string             true  "the ID of the order being updated"
// @Success      200  {object}  OrderStatusResponse
// @Failure      400  {object}  ApiError
// @Failure      404  {object}  ApiError
// @Failure      500  {object}  ApiError
// @Router       /order/{orderId} [post]
func (_Controller *OrderController) UpdateOrderStatus(ctx *gin.Context) {
	var req OrderUpdateRequest
	ctx.BindJSON(&req)

	if strings.EqualFold(req.Status, "delivered") {
		_Controller.deliverOrder(ctx)
	} else if strings.EqualFold(req.Status, "canceled") {
		_Controller.cancelOrder(ctx)
	} else {
		ctx.JSON(400, ApiError{
			Error: "Invalid status. Expected 'delivered' or 'canceled'",
		})
	}
}

// Determines who currently owns the deliver token - the vendor or the customer.
// Returns the owner's address.

// GetTokenOwner godoc
// @Summary      Get the current owner of the delivery contract token
// @Description  Determines who currently owns the deliver token - the vendor or the customer.
// @Description  This looks up the contract in the blockchain rather than reading the status from the database.
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        orderId        path   string    true  "the ID of the order to look up"
// @Success      200  {object}  TokenOwnerResponse
// @Failure      400  {object}  ApiError
// @Failure      404  {object}  ApiError
// @Failure      500  {object}  ApiError
// @Router       /order/{orderId}/owner [get]
func (_Controller *OrderController) GetDeliveryTokenOwner(ctx *gin.Context) {

	orderId := ctx.Param("orderId")
	order, _ := _Controller.OrderRepository.GetOrder(orderId)
	if order == nil {
		orderNotFoundResponse(ctx, orderId)
		return
	}

	owner, err := _Controller.ContractExecutor.GetOwner(order.TokenAddress, _Controller.ServerPrivateKey)

	if err != nil {
		ctx.JSON(500, ApiError{
			Error: err.Error(),
		})
	} else {
		ctx.JSON(200, TokenOwnerResponse{
			Owner: owner,
		})
	}
}

// Delivers the order to the customer. This is represented by transferring the token from the vendor to
// the customer, and transferring Ether from the customer to the vendor to pay for shipping.
func (_Controller *OrderController) deliverOrder(ctx *gin.Context) {
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
		ctx.JSON(500, ApiError{
			Error: err.Error(),
		})
		return
	} else if order == nil || order.Delivered {
		orderNotFoundResponse(ctx, orderId)
		return
	}

	// buy the token from the vendor, thereby accepting delivery of the package
	err = _Controller.ContractExecutor.BuyNFT(order.TokenAddress, order.TokenId, customerPrivateKey, order.Price)
	if err != nil {
		ctx.JSON(500, ApiError{
			Error: err.Error(),
		})
		return
	}

	_Controller.OrderRepository.MarkOrderDelivered(orderId)

	ctx.JSON(200, OrderStatusResponse{
		Status: "delivered",
	})

	return
}

func (_Controller *OrderController) cancelOrder(ctx *gin.Context) {
	orderId := ctx.Param("orderId")
	order, _ := _Controller.OrderRepository.GetOrder(orderId)
	if order == nil {
		orderNotFoundResponse(ctx, orderId)
		return
	}

	err := _Controller.ContractExecutor.BurnContract(order.TokenAddress, _Controller.ServerPrivateKey)
	if err != nil {
		ctx.JSON(500, ApiError{
			Error: "Failed to cancel the order",
		})
	}

	ctx.JSON(200, OrderStatusResponse{
		Status: "canceled",
	})
}

// Ensures all of the query parameters are present in the request
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
		ctx.JSON(400, ApiError{
			Error: sb.String(),
		})
		return false
	}
	return true
}

func orderNotFoundResponse(ctx *gin.Context, orderId string) {
	ctx.JSON(404, ApiError{
		Error: fmt.Sprintf("Order ID [%s] does not exist", orderId),
	})
}
