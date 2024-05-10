package routes

import (
	"carte/controllers"

	"github.com/gin-gonic/gin"
)

func CarteRoute(router *gin.Engine) {
	router.POST("/createcontainer", controllers.CreateContainer())
}
