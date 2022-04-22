package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"learningbay24.de/backend/config"
	"learningbay24.de/backend/course"
	"learningbay24.de/backend/models"
)

func GetCourseById(c *gin.Context) {
	db := config.SetupDbHandle()
	//Get given ID from the Context
	//Convert data type from str to int to use ist as param
	id, err := strconv.Atoi(c.Param("id"))
	//Fetch Data from Database with Backend function
	var course *models.Course
	course, err = models.FindCourse(context.Background(), db, id)
	if err != nil {
		fmt.Println(err.Error())
	}

	c.Header("Access-Control-Allow-Origin", "*")
	//Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, course)
	fmt.Println("course ", course)
	fmt.Println(err)
}

func CreateCourse(c *gin.Context) {
	userid := []int{3}
	db := config.SetupDbHandle()
	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		return
	}
	id, err := course.CreateCourse(db, newCourse.Name, newCourse.EnrollKey, newCourse.Description, userid)
	if err != nil {
		fmt.Println(err.Error())
		panic("error creating course")
	}

	newCourse.ID = id

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, newCourse)
}

func GetCourses(c *gin.Context) {
	db := config.SetupDbHandle()
	var courses []models.Course

	err := queries.Raw("select * from course").Bind(context.Background(), db, &courses)
	if err != nil {
		fmt.Println(err.Error())
		panic("error raw query")
	}

	c.Header("Access-Control-Allow-Origin", "*")
	//Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, courses)
	fmt.Println("courses ", courses)
	fmt.Println(err)
}
