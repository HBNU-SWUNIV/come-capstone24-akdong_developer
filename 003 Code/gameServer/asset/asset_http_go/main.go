package main

import (
	"asset_http_go/configs"
	"asset_http_go/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	fpLog, err := os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fpLog.Close()

	// 표준로거를 파일로그로 변경
	log.SetOutput(fpLog)
	// 데이터베이스 실행
	configs.ConnectDB()

	// 라우트
	routes.AssetRoute(router)

	log.Println("AssetServer is starting...")
	router.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatalf("AssetServer failed to start: %v", err)
	}
}
