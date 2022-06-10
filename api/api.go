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

const (
	AdminRoleId int = iota + 1
	ModeratorRoleId
	UserRoleId
)

type PublicController struct {
	Database *sql.DB
}

func AuthorizeModerator(roleId int) bool {
	log.Infof("Authorizing with role id: %d", roleId)
	return roleId <= ModeratorRoleId
}

func (f *PublicController) GetDataFromCookie(c *gin.Context) (interface{}, error) {
	Cookie := c.Request.Header.Get("Cookie")
	if Cookie == "" {
		log.Errorf("Unable to get cookie")
		c.IndentedJSON(http.StatusBadRequest, "Unable to get cookie")
		return nil, errors.New("Unable to get cookie")
	}
	tokenString := strings.Split(Cookie, "=")[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.Conf.Secrets.JWTSecret), nil
	})
	if err != nil {
		log.Errorf("Error parsing token: %s\n", err.Error())
		return nil, err
	}
	data, ok := token.Claims.(jwt.MapClaims)["data"]
	if !ok {
		return nil, errors.New("Unable to map id from data interface")
	}
	return data, err
}

func (f *PublicController) GetIdFromCookie(c *gin.Context) (int, error) {
	data, err := f.GetDataFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get Data from Cookie: %s\n", err.Error())
		return 0, err
	}
	datamap, ok := data.(map[string]interface{})
	if !ok {
		return 0, errors.New("Unable to map id from data interface")
	}
	id, err := strconv.Atoi(datamap["id"].(string))
	if err != nil {
		log.Errorf("Unable to convert idstring to int: %s\n", err.Error())
		return 0, err
	}
	return id, err

}

func (f *PublicController) GetRoleIdFromCookie(c *gin.Context) (int, error) {
	data, err := f.GetDataFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get Data from Cookie: %s\n", err.Error())
		return 0, err
	}
	datamap, ok := data.(map[string]interface{})
	if !ok {
		return 0, errors.New("Unable to map id from data interface")
	}
	id, err := strconv.Atoi(datamap["role_id"].(string))
	if err != nil {
		log.Errorf("Unable to convert idstring to int: %s\n", err.Error())
		return 0, err
	}
	return id, err
}

func (f *PublicController) GetCourseById(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	// Fetch Data from Database with Backend function
	course, err := course.GetCourse(f.Database, id)
	if err != nil {
		log.Errorf("Unable to get course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, course)
}

func (f *PublicController) DeleteUserFromCourse(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
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
	// Fetch Data from Database with Backend function
	err = course.DeleteUserFromCourse(f.Database, id, user_id)
	if err != nil {
		log.Errorf("Unable to delete user from course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetUsersInCourse(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	// Fetch Data from Database with Backend function
	users, err := course.GetUsersInCourse(f.Database, id)
	if err != nil {
		log.Errorf("Unable to get users in course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, users)
}

func (f *PublicController) GetCoursesFromUser(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Fetch Data from Database with Backend function
	courses, err := course.GetCoursesFromUser(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get courses from user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, courses)
}

func (f *PublicController) DeleteCourse(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	// Deactivate Data from Database with Backend function
	course, err := course.DeleteCourse(f.Database, id)
	// Return Status and Data in JSON-Format
	if err != nil {
		log.Errorf("Unable to delete course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, course)
}

func (f *PublicController) CreateCourse(c *gin.Context) {
	role_id, err := f.GetRoleIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get role_id from Cookie: %s\n", err.Error())
		c.IndentedJSON(http.StatusUnauthorized, err.Error())
		return
	}

	if !AuthorizeModerator(role_id) {
		c.Status(http.StatusUnauthorized)
		return
	}

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

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
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

	c.IndentedJSON(http.StatusOK, newCourse)
}

func (f *PublicController) EnrollUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to string: %s\n", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
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
