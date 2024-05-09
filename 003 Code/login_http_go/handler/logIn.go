package handler

import (
	"database/sql"
	"fmt"
	"net/http"

	"capstone.com/module/db"
	"capstone.com/module/hashing"
	"capstone.com/module/models"
	"github.com/labstack/echo"
)

func LogIn(c echo.Context) error {
	user := new(models.User)

	// model
	type UserInfo struct {
		// DB 고유 아이디 보내기 
		User_id  string `json:"user_id"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
	}

	type Code struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	type CodeInfo struct {
		Code    int      `json:"code"`
		Message UserInfo `json:"message"`
	}

	// 전송 코드
	e_1 := &Code{
		Code:    0,
		Message: "bad request",
	}

	e_2 := &Code{
		Code:    0,
		Message: "user not found",
	}

	e_3 := &Code{
		Code:    0,
		Message: "password mismatch",
	}

	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusOK, e_1)
	}

	// db 연결
	db := db.GetConnector()
	fmt.Println("Connected DB")

	var recv_userID string
	var recv_userPW string
	// inputpw := user.User_pw

	// 가입여부 확인
	err := db.QueryRow("SELECT user_id, user_pw FROM users WHERE user_id = ?", user.User_id).Scan(&recv_userID, &recv_userPW)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusOK, e_2)
	}

	// 비밀번호 검증
	res := hashing.CheckHashPassword(recv_userPW, user.User_pw)
	if !res {
		return c.JSON(http.StatusOK, e_3)
	}

	var send_userId = user.User_id
	var send_Nickname string
	var send_Email string

	err2 := db.QueryRow("SELECT nickname, email FROM users WHERE user_id = ?", recv_userID).Scan(&send_Nickname, &send_Email)
	if err2 != nil {
		return err2
	}

	c_1 := &CodeInfo{
		Code: 1,
		Message: UserInfo{
			User_id:  send_userId,
			Nickname: send_Nickname,
			Email:    send_Email,
		},
	}

	return c.JSON(http.StatusOK, c_1)
}
