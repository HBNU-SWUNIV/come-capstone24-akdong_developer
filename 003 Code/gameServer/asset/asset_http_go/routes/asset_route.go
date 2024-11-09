package routes

import (
	"asset_http_go/controllers"

	"github.com/gin-gonic/gin"
)

func AssetRoute(router *gin.Engine) {
	router.POST("/asset_upload", controllers.CreateAsset())
	// router.POST("/asset_down_valid")
	// 다운로드를 할 때, 이 사용자가 이 에셋을 구매한 이력이 있는지 검증을 한 번 해야함
	router.GET("/asset_search", controllers.SearchAsset())
	router.GET("/asset_down", controllers.DownAsset())
	// 맵을 불러올 때에는 검증 절차가 필요하지 않으니까 지금의 [GET]/asset_down도 있으면 좋음
	router.GET("/asset_info", controllers.GetAsset())
	router.GET("/asset_thumbnail", controllers.DownThumbnail())
}
