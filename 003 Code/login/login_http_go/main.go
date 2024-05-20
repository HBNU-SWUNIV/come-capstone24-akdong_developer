package main

import (
	"log"
	"os"

	"login_http_go/handler"

	"github.com/joho/godotenv"
	"github.com/labstack/echo" // echo프레임워크 사용
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fpLog, err := os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fpLog.Close()

	// 표준로거를 파일로그로 변경
	log.SetOutput(fpLog)

	e := echo.New()
	e.POST("/signup", handler.SignUp)
	e.POST("/login", handler.LogIn)

	e.Logger.Fatal(e.Start("0.0.0.0:8000"))

}
