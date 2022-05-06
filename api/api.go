package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"learningbay24.de/backend/course"
	"learningbay24.de/backend/models"

	"github.com/volatiletech/null/v8"
)

type PublicController struct {
	Database *sql.DB
}

func (f *PublicController) GetCourseById(c *gin.Context) {
	//Get given ID from the Context
	//Convert data type from str to int to use ist as param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	//Fetch Data from Database with Backend function
	course, err := course.GetCourse(f.Database, id)
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
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	//Fetch Data from Database with Backend function
	err = course.DeleteUserFromCourse(f.Database, id, user_id)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Header("Access-Control-Allow-Origin", "*")
	c.Status(http.StatusNoContent)

}

func (f *PublicController) GetUsersInCourse(c *gin.Context) {

	//Get given ID from the Context
	//Convert data type from str to int to use ist as param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	//Fetch Data from Database with Backend function
	users, err := course.GetUsersInCourse(f.Database, id)
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

func (f *PublicController) GetCoursesFromUser(c *gin.Context) {

	//Get given ID from the Context
	//Convert data type from str to int to use ist as param
	user_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	//Fetch Data from Database with Backend function
	courses, err := course.GetCoursesFromUser(f.Database, user_id)
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

func (f *PublicController) DeleteCourse(c *gin.Context) {

	//Get given ID from the Context
	//Convert data type from str to int to use ist as param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	//Deactivate Data from Database with Backend function
	course, err := course.DeleteCourse(f.Database, id)
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

	raw, err := c.GetRawData()
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Printf("Unable to unmarshal the json body: %+v", raw)
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Printf("user_id: %+T", j["user_id"])
	tmp, ok := j["user_id"].(float64)
	if !ok {
		log.Println("unable to convert user_id to float64")
		c.Status(http.StatusInternalServerError)
		return
	}
	user_id := int(tmp)

	name, ok := j["name"].(string)
	if !ok {
		log.Println("unable to convert name to string")
		c.Status(http.StatusInternalServerError)
		return
	}
	description, ok := j["description"].(string)
	if !ok {
		log.Println("unable to convert description to string")
		c.Status(http.StatusInternalServerError)
		return
	}
	enroll_key, ok := j["enroll_key"].(string)
	if !ok {
		log.Println("unable to convert enroll_key to string")
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := course.CreateCourse(f.Database, name, null.StringFrom(description), enroll_key, user_id)
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

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	log.Println(err.Error())
	user_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
	}

	user, err := course.EnrollUser(f.Database, user_id, id, newCourse.EnrollKey)

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	log.Println("user", user)
	c.IndentedJSON(http.StatusOK, newCourse)
}

func (f *PublicController) UpdateCourseById(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	course, err := course.UpdateCourse(f.Database, id, newCourse.Name, newCourse.Description, newCourse.EnrollKey)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	newCourse.ID = id
	log.Println("course", course)
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, newCourse)
}
