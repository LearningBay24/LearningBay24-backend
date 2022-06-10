package api

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"learningbay24.de/backend/exam"
	"time"

	"learningbay24.de/backend/course"
	"learningbay24.de/backend/models"

	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
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

func (f *PublicController) CreateExam(c *gin.Context) {
	var newExam models.Exam

	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Errorf("Unable to unmarshal the json body: %+v", raw)
		c.Status(http.StatusInternalServerError)
		return
	}

	name, ok := j["name"].(string)
	if !ok {
		log.Error("unable to convert name to string")
		c.Status(http.StatusInternalServerError)
		return
	}

	description, ok := j["description"].(string)
	if !ok {
		log.Error("unable to convert description to string")
		c.Status(http.StatusInternalServerError)
		return
	}

	date, err := time.Parse(time.RFC3339, c.Param("date"))
	if err != nil {
		log.Errorf("Unable to convert parameter `date` to time.Time: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	duration, err := strconv.Atoi(c.Param("duration"))
	if err != nil {
		log.Errorf("Unable to convert parameter `duration` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	courseId, err := strconv.Atoi(c.Param("course_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `course_id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	creatorId, err := strconv.Atoi(c.Param("creator_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `creator_id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	online, err := strconv.Atoi(c.Param("online"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	location, ok := j["location"].(string)
	if !ok {
		log.Error("unable to convert location to string")
		c.Status(http.StatusInternalServerError)
		return
	}

	registerDeadline, err := time.Parse(time.RFC3339, c.Param("register_deadline"))
	if err != nil {
		log.Errorf("Unable to convert parameter `deregister_deadline` to time.Time: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	deregisterDeadline, err := time.Parse(time.RFC3339, c.Param("deregister_deadline"))
	if err != nil {
		log.Errorf("Unable to convert parameter `deregister_deadline` to time.Time: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	id, err := pCtrl.CreateExam(name, description, date, duration, courseId, creatorId, int8(online), null.StringFrom(location), null.TimeFrom(registerDeadline), null.TimeFrom(deregisterDeadline))
	if err != nil {
		log.Errorf("Unable to create exam: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	newExam.ID = id

	c.IndentedJSON(http.StatusOK, newExam)
}
