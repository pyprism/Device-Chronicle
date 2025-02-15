package controllers

import "github.com/gin-gonic/gin"

type StaticFileController struct{}

func (n *StaticFileController) StaticFile(context *gin.Context) {
	static := context.Param("static")
	context.File("./static/" + static)
}
