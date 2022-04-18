package api

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"learningbay24.de/backend/course"
	"learningbay24.de/backend/models"
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
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", Conf.DB.User, Conf.DB.Pass, Conf.DB.Host, Conf.DB.Port, Conf.DB.Database)
}

func GetCourseById(c *gin.Context) {
    initConfig()
    db := setupDbHandle()
	//Get given ID from the Context
	//Convert data type from str to int to use ist as param 
	id,err := strconv.Atoi(c.Param("id"))
	//Fetch Data from Database with Backend function
    course := course.GetCourse(db,id)
	//Return Status and Data in JSON-Format
    c.IndentedJSON(http.StatusOK, course)
	fmt.Println(err)
}

func CreateCourse(c *gin.Context){
    userid := []int{1,}
    initConfig()
    db := setupDbHandle()
    var newCourse models.Course
    if err := c.BindJSON(&newCourse); err != nil{
        return
    }
    course.CreateCourse(db,newCourse.Name,newCourse.EnrollKey,newCourse.Description,userid)
    c.IndentedJSON(http.StatusOK, newCourse)
}
