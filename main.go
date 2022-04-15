package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"

	"learningbay24.de/backend/config"
)

func setupDbHandle() *sql.DB {
	dsn := getDataSourceName()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func getDataSourceName() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Conf.DB.User, config.Conf.DB.Pass, config.Conf.DB.Host, config.Conf.DB.Port, config.Conf.DB.Database)
}

func main() {
	config.InitConfig()
	db := setupDbHandle()
}
