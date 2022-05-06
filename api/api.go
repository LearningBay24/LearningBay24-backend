package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"learningbay24.de/backend/course"
	"learningbay24.de/backend/models"
)

type PublicController struct {
	Database *sql.DB
}


func (f *PublicController) GetCourseById(c *gin.Context) {
	//Get given ID from the Context
	//Convert data type from str to int to use ist as param 
	id,err := strconv.Atoi(c.Param("id"))
	//Fetch Data from Database with Backend function
    course,err := course.GetCourse(f.Database,id)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
    c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, course)
	log.Println("course ", course)
	log.Println(err)
}


func (f *PublicController) DeleteUserFromCourse(c *gin.Context) {
	//Get given ID from the Context
	//Convert data type from str to int to use ist as param 
	user_id, err := strconv.Atoi(c.Param("user_id"))
	id,err := strconv.Atoi(c.Param("id"))
	//Fetch Data from Database with Backend function
    user,err := course.DeleteUserFromCourse(f.Database,id,user_id)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return 
	}
	//Return Status and Data in JSON-Format
    c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, user)
	log.Println("course ", user)
}


func (f *PublicController) GetUsersInCourse(c *gin.Context) {

	//Get given ID from the Context
	//Convert data type from str to int to use ist as param 
	id,err := strconv.Atoi(c.Param("id"))
	//Fetch Data from Database with Backend function
    users,err := course.GetUsersInCourse(f.Database,id)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return 
	}
	//Return Status and Data in JSON-Format
    c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, users)
	log.Println("course ", users)
}



func (f *PublicController) GetUserCourses(c *gin.Context) {
 
	//Get given ID from the Context
	//Convert data type from str to int to use ist as param 
	user_id, err := strconv.Atoi(c.Param("user_id"))
	//Fetch Data from Database with Backend function
    courses,err := course.GetUserCourses(f.Database,user_id)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return 
	}
	//Return Status and Data in JSON-Format
    c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, courses)
	log.Println("course ", courses)
	
}






func (f *PublicController) DeactivateCourse(c *gin.Context) {

	//Get given ID from the Context
	//Convert data type from str to int to use ist as param 
	id,err := strconv.Atoi(c.Param("id"))
	//Deactivate Data from Database with Backend function
    course,err := course.DeactivateCourse(f.Database,id)
	//Return Status and Data in JSON-Format
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return 
	}
    c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, course)
	log.Println("course ", course)
}




func (f *PublicController) CreateCourse(c *gin.Context) {

	var newCourse models.Course
	//user_id, err := strconv.Atoi(c.Param("user_id"))
	user_id := []int{1,}
	if err := c.BindJSON(&newCourse); err != nil {
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return 
		}
	}
	id,err := course.CreateCourse(f.Database, newCourse.Name,newCourse.Description, newCourse.EnrollKey,  user_id)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return 
	}
	newCourse.ID = id
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, newCourse)
}


func (f *PublicController) EnrollUser(c *gin.Context) {

	
	id,err := strconv.Atoi(c.Param("id"))
	log.Println(err.Error())
	user_id, err := strconv.Atoi(c.Param("user_id"))
	var newCourse models.Course
	
	user,err := course.EnrollUser(f.Database, user_id,id,newCourse.EnrollKey)

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return 
	}

	c.Header("Access-Control-Allow-Origin", "*")
	log.Println("user",user)
	c.IndentedJSON(http.StatusOK, newCourse)
}




func (f *PublicController) UpdateCourseById(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return 
	}
	course,err := course.UpdateCourse(f.Database,id, newCourse.Name,newCourse.Description, newCourse.EnrollKey)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return 
	}
	newCourse.ID = id
	log.Println("course",course)
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, newCourse)
}


/* func GetCourses(c *gin.Context) {
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
} */