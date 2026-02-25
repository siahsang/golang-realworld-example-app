package handler

import "github.com/gin-gonic/gin"

func HandleResponse(ctx *gin.Context, data any, err error) {

	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, data)
	return
}

func BindAndCheck(ctx *gin.Context, ob any) bool {
	if err := ctx.ShouldBind(ob); err != nil {
		return false
	}

	return nil
}
