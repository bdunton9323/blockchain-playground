package controllers

import (
	"fmt"
	"github.com/bdunton9323/blockchain-playground/contract"
	"github.com/gin-gonic/gin"
)

func DeployContract(ctx *gin.Context) {
	fmt.Println("Contract ID:", contract.InstallContract())
	ctx.String(200, contract.InstallContract())
}
