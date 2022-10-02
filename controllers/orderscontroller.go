package controllers

import (
	"strconv"

	"github.com/bdunton9323/blockchain-playground/contract"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type OrderController struct {
	NodeUrl          *string
	ServerPrivateKey *string
}

// Creates an order that can later be delivered
func (_Controller *OrderController) CreateOrder(ctx *gin.Context) {

	price, err := strconv.ParseInt(ctx.Query("price"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": "Price value must be integer",
		})
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

	// TODO: store the contract, tokenId, and orderId in the database
	orderId := uuid.New()

	ctx.JSON(200, gin.H{
		"address": address,
		"tokenId": tokenId,
		"orderId": orderId.String(),
	})
}

func (_Controller *OrderController) DeliverOrder(ctx *gin.Context) {
	orderId := ctx.Param("orderId")

	// the customer is signing for the order, and they have a different key than
	// the one loaded into the server. I recognize that transmitting the private
	// key to a server is not a good idea, but with a lack of time it demonstrates
	// the functionality.
	customerPrivateKey := ctx.Query("customerKey")

	log.Infof("Delivering order [%v]", orderId)

	// TODO: get the address and tokenId from the database instead of the query
	contractAddr := ctx.Query("address")
	tokenId, _ := strconv.ParseInt(ctx.Query("tokenId"), 10, 32)

	err := contract.BuyNFT(contractAddr, tokenId, customerPrivateKey, *_Controller.NodeUrl)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
	} else {
		ctx.JSON(200, gin.H{
			"status": "delivered",
		})
	}
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
