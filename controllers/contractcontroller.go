package controllers

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/bdunton9323/blockchain-playground/contract"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// TODO: inject the node URL from above somewhere. Private key can also be injected.
const nodeUrl = "http://172.13.3.1:8545"

func DeployNFTContract(ctx *gin.Context, privateKey string) {

	i, err := strconv.ParseInt(ctx.Query("price"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": "Price value must be integer",
		})
	}
	tokenId, address, err := contract.MintNFT(privateKey, nodeUrl, i)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
	}

	// TODO: generate an order ID and associate it with the contract address

	ctx.JSON(200, gin.H{
		"address": address,
		"tokenId": tokenId,
	})
}

func BuyNFT(ctx *gin.Context, orderId string, privateKey string) {
	orderId = ctx.Query("orderId")
	// TODO: get the address from the database instead of the query
	contractAddr := ctx.Query("address")

	contract.BuyNFT(contractAddr, privateKey, nodeUrl)
}

// Old stuff below:

func DeployContract(ctx *gin.Context, privateKey string) {
	contractAddress, err := contract.InstallContract(privateKey)
	if err != nil {
		ctx.String(500, err.Error())
	} else {
		log.Infof("Contract address: %v", contractAddress)
		ctx.String(200, contractAddress)
	}
}

func ExecuteContract(ctx *gin.Context, privateKey string) {
	newValue := new(big.Int)
	newValue, ok := newValue.SetString(ctx.Query("value"), 10)
	if !ok {
		ctx.JSON(400, gin.H{
			"error": fmt.Sprintf("Invalid value: %s", ctx.Query("value")),
		})
	}
	contract.SetValue(newValue, ctx.Query("address"), privateKey)
	ctx.JSON(200, gin.H{
		"value": fmt.Sprintf("%v", newValue),
	})
}

func GetContractValue(ctx *gin.Context, privateKey string) {
	val, err := contract.GetValue(ctx.Query("address"))
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
	} else {
		ctx.JSON(200, gin.H{
			"value": &val,
		})
	}
}
