package main

import (
	"carte/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// 라우트
	routes.CarteRoute(router)

	router.Run("")
}
