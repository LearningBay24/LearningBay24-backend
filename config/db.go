package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func SetupDbHandle() *sql.DB {
	dsn := getDataSourceName()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func getDataSourceName() string {

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", Conf.DB.User, Conf.DB.Pass, Conf.DB.Host, Conf.DB.Port, Conf.DB.Database)

}
