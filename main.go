package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml"
	"learningbay24.de/backend/api"
)

type DB struct {
	Host     string
	Port     int16
	User     string
	Pass     string
	Database string
}

type Config struct {
	DB DB
}

var (
	Conf Config
)

func initConfig() {
	data, err := ioutil.ReadFile("./config.toml")
	if err != nil {
		log.Fatal(err)
	}
	err = toml.Unmarshal(data, &Conf)
	if err != nil {
		log.Fatal(err)
	}
}

func setupDbHandle() *sql.DB {
	dsn := getDataSourceName()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func getDataSourceName() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", Conf.DB.User, Conf.DB.Pass, Conf.DB.Host, Conf.DB.Port, Conf.DB.Database)
}

func main() {
	router := gin.Default()
	router.GET("/courses/:id",api.GetCourseById)
	router.POST("/courses",api.CreateCourse)
	router.Run("localhost:8080")
}
