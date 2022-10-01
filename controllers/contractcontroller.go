package controllers

import (
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
