package routes

import (
	"capstone.com/controllers"
	"github.com/gin-gonic/gin"
)

func MapRoute(router *gin.Engine) {
	// map 관련 모든 routes 관리
	router.POST("/map_data", controllers.CreateMap())
	router.GET("/map_data", controllers.GetMap())
	router.GET("/maplist", controllers.GetList())
}
