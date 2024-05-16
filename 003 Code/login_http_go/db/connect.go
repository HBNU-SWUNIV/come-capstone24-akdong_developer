package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func GetConnector() *sql.DB {
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "localhost:3306",
		Collation:            "utf8mb4_general_ci",
		Loc:                  time.UTC,
		MaxAllowedPacket:     4 << 20.,
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		DBName:               os.Getenv("DBNAME"),
	}
	connector, err := mysql.NewConnector(&cfg)
	if err != nil {
		panic(err)
	}
	db := sql.OpenDB(connector)

	err = db.Ping()

	if err != nil {
		fmt.Println("Could not connect to databases:", err)
	}

	// Database 없는 경우
	_, err = db.Exec("CREATE DATABASES IF NOT EXISTS go_signup")
	if err != nil {
		fmt.Println("Could not create databases:", err)
	}

	// 테이블 생성
	err = createTable(db)
	if err != nil {
		fmt.Println("Could not create table:", err)
	}

	return db

}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF EXISTS users (
			ID int NOT NULL AUTO_INCREMENT PRIMARY KEY,
			user_id varchar(30) UNIQUE,
			user_pw varchar(1000),
			nickname varchar(30),
			email varchar(50) NOT NULL
		) 
	`)
	if err != nil {
		return err
	}
	return nil
}
