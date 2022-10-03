package controllers

import (
	"fmt"

	"github.com/bdunton9323/blockchain-playground/contract"
	"github.com/bdunton9323/blockchain-playground/orders"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type OrderController struct {
	NodeUrl          *string
	ServerPrivateKey *string
	OrderRepository  *orders.MariaDBOrderRepository
}

// Creates an order that can later be delivered
func (_Controller *OrderController) CreateOrder(ctx *gin.Context) {

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
	price := int64(5)
	tokenId, address, err := contract.DeployContractAndMintNFT(
		*_Controller.ServerPrivateKey,
		*_Controller.NodeUrl,
		price,
		userAddress)

	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
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
		ctx.JSON(500, gin.H{
			"error": "Error writing order to database",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"address":         address,
		"tokenId":         tokenId,
		"contractAddress": address,
		"orderId":         order.OrderId,
	})
}

func (_Controller *OrderController) DeliverOrder(ctx *gin.Context) error {
	// The customer is signing for the order, and they have a different key than
	// the one loaded into the server. I recognize that transmitting the private
	// key to a server is a terrible idea, but given the lack of time, it at least
	// demonstrates the smart contract's functionality.
	customerPrivateKey := ctx.Query("customerKey")

	orderId := ctx.Param("orderId")
	log.Infof("Delivering order [%v]", orderId)

	order, err := _Controller.OrderRepository.GetOrder(orderId)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return err
	} else if order == nil || order.Delivered {
		ctx.JSON(400, gin.H{
			"error": "Could not find that order ID. Has it already been delivered?",
		})
		return nil
	}

	// buy the token from the vendor, thereby accepting delivery of the package
	err = contract.BuyNFT(order.TokenAddress, order.TokenId, customerPrivateKey, *_Controller.NodeUrl)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return err
	}

	_Controller.OrderRepository.MarkOrderDelivered(orderId)

	ctx.JSON(200, gin.H{
		"status": "delivered",
	})

	return nil
}

// Determines who currently owns the deliver token - the vendor or the customer.
// Returns the owner's address.
func (_Controller *OrderController) GetDeliveryTokenOwner(ctx *gin.Context) {
	orderId := ctx.Param("orderId")
	order, err := _Controller.OrderRepository.GetOrder(orderId)
	if err != nil || order == nil {
		ctx.JSON(400, gin.H{
			"error": fmt.Sprintf("Order ID [%s] does not exist", orderId),
		})
	}

	owner, err := contract.GetOwner(order.TokenAddress, *_Controller.ServerPrivateKey)

	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
	} else {
		ctx.JSON(200, gin.H{
			"owner": owner,
		})
	}
}
