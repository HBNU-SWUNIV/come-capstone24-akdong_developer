package main

import (
	"log"
	"map_http_go/configs"
	"map_http_go/routes"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	fpLog, err := os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fpLog.Close()

	log.SetOutput(fpLog)

	router := gin.Default()

	configs.ConnectDB()

	routes.MapRoute(router)

	router.Run("0.0.0.0:8070")

}
