package controllers

import (
	"strconv"

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

	price, err := strconv.ParseInt(ctx.Query("price"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": "Price value must be integer",
		})
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
	userAddress := ctx.Query("address")

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
	} else if order == nil {
		ctx.JSON(400, gin.H{
			"error": "Could not find that order ID",
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

	ctx.JSON(200, gin.H{
		"status": "delivered",
	})

	return nil
}

func (_Controller *OrderController) GetOrderStatus(ctx *gin.Context) {
	// TODO: this should only take in the order ID and look up the contract in the DB
	//orderId := ctx.Param("orderId")
	contractAddress := ctx.Query("contract")
	owner, err := contract.GetOwner(contractAddress, *_Controller.ServerPrivateKey)

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
