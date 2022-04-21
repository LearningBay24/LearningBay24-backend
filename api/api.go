package api

import (
	"fmt"
	"net/http"
	"strconv"
	"context"

	"github.com/gin-gonic/gin"
	"learningbay24.de/backend/course"
	"learningbay24.de/backend/models"
	"learningbay24.de/backend/config"
)

func GetCourseById(c *gin.Context) {
    db := config.SetupDbHandle()
	//Get given ID from the Context
	//Convert data type from str to int to use ist as param 
	id,err := strconv.Atoi(c.Param("id"))
	//Fetch Data from Database with Backend function
	var course *models.Course
	course, err = models.FindCourse(context.Background(), db, id)
	if err != nil {
		panic("nah")
	}

	//Return Status and Data in JSON-Format
    c.IndentedJSON(http.StatusOK, course)
	fmt.Println(err)
}

func CreateCourse(c *gin.Context){
    userid := []int{1,}
    db := config.SetupDbHandle()
    var newCourse models.Course
    if err := c.BindJSON(&newCourse); err != nil{
        return
    }
    course.CreateCourse(db,newCourse.Name,newCourse.EnrollKey,newCourse.Description,userid)
    c.IndentedJSON(http.StatusOK, newCourse)
}
