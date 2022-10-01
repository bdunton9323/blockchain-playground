package controllers

import (
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
	contract.SetValue(big.NewInt(87), ctx.Query("address"), privateKey)
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
