package models

type User struct{
	User_id string `json:user_id`
	User_pw string `json:user_pw`
	Nickname string `json:nickname`
	Email string `json:email`
}

