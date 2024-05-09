package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"capstone.com/module/db"
	"capstone.com/module/hashing"
	"capstone.com/module/models"
	"github.com/labstack/echo"
)

func SignUp(c echo.Context) error {
	user := new(models.User)

	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"bad request": "0",
		})
	}

	db := db.GetConnector()
	fmt.Println("Connected DB")

	// 아이디 존재 여부 확인(아이디 중복 방지를 위함)
	query_id := fmt.Sprintf("SELECT * FROM users WHERE user_id ='%s';", user.User_id)
	fmt.Println(query_id)
	// _, err := db.Exec(query_id) // Exec는 insert, update, delete하기 위해 사용
	result_id := db.QueryRow(query_id).Scan(&user.User_id)
	if result_id != sql.ErrNoRows {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"existing id": "0",
		})
	}

	// 닉네임 존재 여부 확인
	query_nick := fmt.Sprintf("SELECT * FROM users WHERE nickname ='%s';", user.Nickname)
	fmt.Println(query_nick)
	result_nick := db.QueryRow(query_nick).Scan(&user.Nickname)
	if result_nick != sql.ErrNoRows {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"existing nickname": "0",
		})
	}

	// 비밀번호 bycrypt 라이브러리 해싱 처리  ---> (무슨 해싱?)
	hashpw, err := hashing.HashPassword(user.User_pw)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}
	user.User_pw = hashpw

	// 유저 생성
	query_r := "INSERT INTO users (user_id, user_pw, nickname, email) VALUES (?, ?, ?, ?)"
	// fmt.Println(query_r)
	_, err = db.Exec(query_r, user.User_id, user.User_pw, user.Nickname, user.Email)
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	// Success
	return c.JSON(http.StatusOK, map[string]string{
		"Success": "1",
	})

}

