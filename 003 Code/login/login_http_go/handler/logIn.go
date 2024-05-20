package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"login_http_go/db"
	"login_http_go/hashing"
	"login_http_go/models"

	"github.com/labstack/echo"
)

func LogIn(c echo.Context) error {
	user := new(models.User)

	// ---------- 전송 코드 -------------------
	e_1 := &models.Code{
		Code:    0,
		Message: "bad request",
	}

	e_2 := &models.Code{
		Code:    0,
		Message: "user not found",
	}

	e_3 := &models.Code{
		Code:    0,
		Message: "password mismatch",
	}
	// -----------------------------------------

	if err := c.Bind(user); err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, e_1)
	}

	// db 연결
	db := db.GetConnector()
	fmt.Println("Connected DB")

	var recv_userID string
	var recv_userPW string

	// 가입여부 확인
	err := db.QueryRow("SELECT user_id, user_pw FROM users WHERE user_id = ?", user.User_id).Scan(&recv_userID, &recv_userPW)
	if err == sql.ErrNoRows {
		log.Println(err)
		return c.JSON(http.StatusOK, e_2)
	}
	log.Println() // 로그 출력()

	// 비밀번호 검증
	res := hashing.CheckHashPassword(recv_userPW, user.User_pw)
	if !res {
		log.Println(err)
		return c.JSON(http.StatusOK, e_3)
	}

	var send_userId = user.User_id
	var send_Nickname string
	var send_Email string

	// 로그인 성공후 반환
	err = db.QueryRow("SELECT nickname, email FROM users WHERE user_id = ?", recv_userID).Scan(&send_Nickname, &send_Email)
	if err != nil {
		log.Println(err)
		return err
	}

	c_1 := &models.CodeInfo{
		Code: 1,
		Message: models.UserInfo{
			User_id:  send_userId,
			Nickname: send_Nickname,
			Email:    send_Email,
		},
	}

	return c.JSON(http.StatusOK, c_1)
}
