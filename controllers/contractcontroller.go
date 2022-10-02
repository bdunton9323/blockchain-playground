package controllers

import (
	"fmt"
	"math/big"

	"github.com/bdunton9323/blockchain-playground/contract"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

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
