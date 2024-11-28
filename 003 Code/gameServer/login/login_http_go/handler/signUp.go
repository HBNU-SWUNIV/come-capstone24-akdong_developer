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

func SignUp(c echo.Context) error {
	user := new(models.User)

	// 회원가입 정보 받아오기
	if err := c.Bind(user); err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, &models.Code{
			Code: 0, Message: "bad request",
		})
	}

	// db 연결
	db := db.GetConnector()
	log.Println("Connected DB")

	// 아이디 존재 여부 확인(아이디 중복 방지를 위함)
	query_id := fmt.Sprintf("SELECT * FROM users WHERE user_id ='%s';", user.User_id)

	// _, err := db.Exec(query_id) // Exec는 insert, update, delete하기 위해 사용
	err := db.QueryRow(query_id).Scan(&user.User_id)
	if err != sql.ErrNoRows {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, &models.Code{
			Code: 0, Message: "existing id",
		})
	}

	// 비밀번호 bycrypt 라이브러리 해싱 처리
	hashpw, err := hashing.HashPassword(user.User_pw)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, &models.Code{
			Code: 0, Message: "hashing err",
		})
	}
	user.User_pw = hashpw

	// 유저 생성
	query_r := "INSERT INTO users (user_id, user_pw, nickname, email) VALUES (?, ?, ?, ?)"

	_, err = db.Exec(query_r, user.User_id, user.User_pw, user.Nickname, user.Email)
	if err != nil {
		log.Println("Failed to insert data: %v", err)
	}

	// Success
	return c.JSON(http.StatusOK, &models.Code{
		Code: 1, Message: "Sign up Success",
	})
}

// // 닉네임 존재 여부 확인 --> 존재 하는 걸로 변경
// query_nick := fmt.Sprintf("SELECT * FROM users WHERE nickname ='%s';", user.Nickname)
// fmt.Println(query_nick)
// result_nick := db.QueryRow(query_nick).Scan(&user.Nickname)
// if result_nick != sql.ErrNoRows {
// 	return c.JSON(http.StatusBadRequest, map[string]string{
// 		"existing nickname": "0",
// 	})
// }
