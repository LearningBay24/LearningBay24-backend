package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"learningbay24.de/backend/calender"
	"learningbay24.de/backend/config"
	"learningbay24.de/backend/course"
	"learningbay24.de/backend/dbi"
	"learningbay24.de/backend/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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
		log.Errorf("Unable to get course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, course)
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
		log.Errorf("Unable to delete user from course: %s\n", err.Error())
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
		log.Errorf("Unable to get users in course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, users)
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
		log.Errorf("Unable to get courses from user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, courses)

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
		log.Errorf("Unable to delete course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, course)
}

func (f *PublicController) CreateCourse(c *gin.Context) {

	var newCourse models.Course

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

	tmp, ok := j["user_id"].(float64)
	if !ok {
		log.Error("unable to convert user_id to float64")
		c.Status(http.StatusInternalServerError)
		return
	}
	user_id := int(tmp)

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
	enroll_key, ok := j["enroll_key"].(string)
	if !ok {
		log.Error("unable to convert enroll_key to string")
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := course.CreateCourse(f.Database, name, null.StringFrom(description), enroll_key, user_id)
	if err != nil {
		log.Errorf("Unable to create course: %s\n", err.Error())
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
		log.Errorf("Unable to convert parameter `id` to string: %s\n", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	user_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		if err != nil {
			log.Errorf("Unable to bind json: %s\n", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
	}

	_, err = course.EnrollUser(f.Database, user_id, id, newCourse.EnrollKey)
	if err != nil {
		log.Errorf("Unable to enroll user in course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
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
		log.Errorf("Unable to bind json: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	_, err = course.UpdateCourse(f.Database, id, newCourse.Name, newCourse.Description, newCourse.EnrollKey)
	if err != nil {
		log.Errorf("Unable to update course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	newCourse.ID = id
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, newCourse)
}

func (f *PublicController) Login(c *gin.Context) {
	//Map the given user on json
	var newUser models.User
	if err := c.BindJSON(&newUser); err != nil {
		log.Errorf("Unable to bind json: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	//Check if credentials of given user are valid
	id, err := dbi.VerifyCredentials(f.Database, newUser.Email, []byte(newUser.Password))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Unable to find user with E-Mail: %s", newUser.Email))
		} else {
			c.IndentedJSON(http.StatusUnauthorized, err.Error())
			log.Errorf("Unable to verify credentials: %s\n", err.Error())
		}

		return
	}

	//Put new Claim on given user
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	//Get signed token with the sercret key
	token, err := claims.SignedString([]byte(config.Conf.Secrets.JWTSecret))
	if err != nil {
		log.Errorf("Unable to get signed token: %s\n", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	//Set the cookie and add it to the response header
	c.SetCookie("user_token", token, int((time.Hour * 24).Seconds()), "/", config.Conf.Domain, config.Conf.Secure, true)
	//Return user with set cookie
	newUser.Password = nil
	newUser.ID = id
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, newUser)
}

func (f *PublicController) Register(c *gin.Context) {
	var newUser models.User
	if err := c.BindJSON(&newUser); err != nil {
		log.Errorf("Unable to bind json: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	id, err := dbi.CreateUser(f.Database, newUser)
	if err != nil {
		log.Errorf("Unable to create user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	}

	newUser.ID = id
	newUser.Password = nil
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusCreated, newUser)
}

func (f *PublicController) GetAllAppointments(c *gin.Context) {

	user_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	appointments, err := calender.GetAllAppointments(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get appointments from user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, appointments)
}

func (f *PublicController) GetAppointments(c *gin.Context) {

	user_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

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
	beforeDate, ok := j["beforeDate"].(time.Time)
	if !ok {
		log.Error("unable to convert beforeDate to time.Time")
		c.Status(http.StatusInternalServerError)
		return
	}
	afterDate, ok := j["afterDate"].(time.Time)
	if !ok {
		log.Error("unable to convert afterDate to time.Time")
		c.Status(http.StatusInternalServerError)
		return
	}

	appointments, err := calender.GetAppointments(f.Database, user_id, beforeDate, afterDate)
	if err != nil {
		log.Errorf("Unable to get appointments from user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, appointments)
}

func (f *PublicController) GetAllSubmissions(c *gin.Context) {

	user_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	appointments, err := calender.GetAllSubmissions(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get submissions from user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, appointments)
}

func (f *PublicController) AddCourseToCalender(c *gin.Context) {

	var newAppointment models.Appointment

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

	date, ok := j["date"].(time.Time)
	if !ok {
		log.Error("unable to convert date to time.Time")
		c.Status(http.StatusInternalServerError)
		return
	}
	location, ok := j["location"].(null.String)
	if !ok {
		log.Error("unable to convert location to null.String")
		c.Status(http.StatusInternalServerError)
		return
	}
	online, ok := j["online"].(int8)
	if !ok {
		log.Error("unable to convert online to int8")
		c.Status(http.StatusInternalServerError)
		return
	}
	courseId, ok := j["courseId"].(int)
	if !ok {
		log.Error("unable to convert courseId to int")
		c.Status(http.StatusInternalServerError)
		return
	}
	repeats, ok := j["repeats"].(bool)
	if !ok {
		log.Error("unable to convert repeats to bool")
		c.Status(http.StatusInternalServerError)
		return
	}
	repeatDistance, ok := j["repeatDistance"].(int)
	if !ok {
		log.Error("unable to convert repeatDistance to int")
		c.Status(http.StatusInternalServerError)
		return
	}
	repeatEnd, ok := j["repeatEnd"].(time.Time)
	if !ok {
		log.Error("unable to convert repeatEnd to time.Time")
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := calender.AddCourseToCalender(f.Database, date, location, online, courseId, repeats, repeatDistance, repeatEnd)
	if err != nil {
		log.Errorf("Unable to create appointment: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	newAppointment.ID = id
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, newAppointment)
}

func (f *PublicController) AddSubmissionToCalender(c *gin.Context) {

	var newAppointment models.Appointment

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

	submDate, ok := j["submDate"].(time.Time)
	if !ok {
		log.Error("unable to convert submDate to time.Time")
		c.Status(http.StatusInternalServerError)
		return
	}
	submName, ok := j["submName"].(null.String)
	if !ok {
		log.Error("unable to convert submName to null.String")
		c.Status(http.StatusInternalServerError)
		return
	}
	courseId, ok := j["courseId"].(int)
	if !ok {
		log.Error("unable to convert courseId to int")
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := calender.AddSubmissionToCalender(f.Database, submDate, submName, courseId)
	if err != nil {
		log.Errorf("Unable to add appointment of submission: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	newAppointment.ID = id
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, newAppointment)
}

func (f *PublicController) DeactivateAppointment(c *gin.Context) {
	// Get given ID from the Context
	//Convert data type from str to int; bool to use ist as param
	appointment_id, err := strconv.Atoi(c.Param("appointment_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	//Fetch Data from Database with Backend function
	err = calender.DeactivateAppointment(f.Database, appointment_id)
	if err != nil {
		log.Errorf("Unable to delete appointment: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Header("Access-Control-Allow-Origin", "*")
	c.Status(http.StatusNoContent)
}

func (f *PublicController) DeactivateCourseInCalender(c *gin.Context) {
	// Get given ID from the Context
	//Convert data type from str to int; bool to use ist as param
	appointment_id, err := strconv.Atoi(c.Param("appointment_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	course_id, err := strconv.Atoi(c.Param("course_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	repeats, err := strconv.ParseBool(c.Param("repeats"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	//Fetch Data from Database with Backend function
	err = calender.DeactivateCourseInCalender(f.Database, appointment_id, course_id, repeats)
	if err != nil {
		log.Errorf("Unable to delete appointment from course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Header("Access-Control-Allow-Origin", "*")
	c.Status(http.StatusNoContent)
}

func (f *PublicController) DeactivateExamInCalender(c *gin.Context) {
	// Get given ID from the Context
	//Convert data type from str to int; bool to use ist as param
	appointment_id, err := strconv.Atoi(c.Param("appointment_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	exam_id, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	//Fetch Data from Database with Backend function
	err = calender.DeactivateExamInCalender(f.Database, appointment_id, exam_id)
	if err != nil {
		log.Errorf("Unable to delete exam from calender: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Header("Access-Control-Allow-Origin", "*")
	c.Status(http.StatusNoContent)
}
