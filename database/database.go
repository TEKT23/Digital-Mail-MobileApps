package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DatabaseInit() {
	var err error
	const MYSQL = "root:@tcp(127.0.0.1:3306)/db_TA?charset=utf8mb4&parseTime=True&loc=Local"
	DSN := MYSQL
	DB, err = gorm.Open(mysql.Open(DSN), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	} else {
		fmt.Println("Connected to database")
	}
}
