package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func respondWithError(ctx *gin.Context, statusCode int, errorString string, err error) {
	ctx.JSON(statusCode, gin.H{
		"Error": fmt.Sprintf("%s: %v\n", errorString, err),
	})
}
