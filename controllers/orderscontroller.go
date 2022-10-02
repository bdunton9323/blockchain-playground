package controllers

import (
	"strconv"

	"github.com/bdunton9323/blockchain-playground/contract"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// TODO: inject the node URL from above somewhere. Private key can also be injected.
const nodeUrl = "http://172.13.3.1:8545"

// Creates an order that can later be delivered
func CreateOrder(ctx *gin.Context, privateKey string) {

	price, err := strconv.ParseInt(ctx.Query("price"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": "Price value must be integer",
		})
	}

	// the user who is allowed to receive the shipment
	userAddress := ctx.Query("address")

	tokenId, address, err := contract.DeployContractAndMintNFT(privateKey, nodeUrl, price, userAddress)
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

func DeliverOrder(ctx *gin.Context) {
	orderId := ctx.Param("orderId")

	// the customer is signing for the order, and they have a different key than
	// the one loaded into the server. I recognize that transmitting the private
	// key to a server is not a good idea, but with a lack of time it demonstrates
	// the functionality.
	customerPrivateKey := ctx.Query("customerKey")

	log.Infof("Delivering order [%v]", orderId)

	// TODO: get the address from the database instead of the query
	contractAddr := ctx.Query("address")

	err := contract.BuyNFT(contractAddr, customerPrivateKey, nodeUrl)
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

func GetOrderStatus(ctx *gin.Context, privateKey string) {
	// TODO: this should only take in the order ID
	//orderId := ctx.Param("orderId")
	contractAddress := ctx.Query("contract")
	owner, err := contract.GetOwner(contractAddress, privateKey)

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
