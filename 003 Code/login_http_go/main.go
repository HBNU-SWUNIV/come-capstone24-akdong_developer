package main

import (
	"log"

	"module/bin"

	"capstone.com/module/handler"
	"github.com/joho/godotenv"
	"github.com/labstack/echo" // echo프레임워크 사용
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	e := echo.New()
	e.POST("/signup", handler.SignUp)
	e.POST("/login", handler.LogIn)

	e.Logger.Fatal(e.Start(bin.IP))

}
