package models

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
