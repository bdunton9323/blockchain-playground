package controllers

import (
	"github.com/gin-gonic/gin"
)

func HelloController(ctx *gin.Context) {
	ctx.String(200, "Hello world!");
}
