package routes

import (
	"carte/controllers"

	"github.com/gin-gonic/gin"
)

func CarteRoute(router *gin.Engine) {
	router.GET("/createcontainer", controllers.CreateContainer)
	//router.POST("/buildimage", controllers.BuildImage)
}
