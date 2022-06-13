package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"learningbay24.de/backend/calender"
	"learningbay24.de/backend/config"
	"learningbay24.de/backend/course"
	coursematerial "learningbay24.de/backend/courseMaterial"
	"learningbay24.de/backend/dbi"
	"learningbay24.de/backend/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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

func (f *PublicController) Login(c *gin.Context) {
	type User struct {
		Firstname           string `json:"firstname"`
		Surname             string `json:"surname"`
		Email               string `json:"email"`
		Password            string `json:"password"`
		RoleID              int    `json:"role_id"`
		PreferredLanguageID int    `json:"preferred_language_id"`
	}

	var tmpUser User
	if err := c.BindJSON(&tmpUser); err != nil {
		log.Errorf("Unable to bind json: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	pw := []byte(tmpUser.Password)
	newUser := models.User{
		Firstname:           tmpUser.Firstname,
		Surname:             tmpUser.Surname,
		Email:               tmpUser.Email,
		Password:            pw,
		RoleID:              tmpUser.RoleID,
		PreferredLanguageID: tmpUser.PreferredLanguageID,
	}

	// Check if credentials of given user are valid
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

	user, err := dbi.GetUserById(f.Database, id)
	if err != nil {
		log.Errorf("Unable to get user by id: %s", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	claims := &jwt.MapClaims{
		"IssuedAt":  time.Now().Unix(),
		"ExpiresAt": time.Now().Add(time.Hour * 24).Unix(),
		"data": map[string]string{
			"id":      strconv.Itoa(id),
			"role_id": strconv.Itoa(user.RoleID),
		},
	}

	// Get signed token with the sercret key
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims)

	secretKey := config.Conf.Secrets.JWTSecret

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Errorf("Unable to sign token: %s\n", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Set the cookie and add it to the response header
	c.SetCookie("user_token", tokenString, int((time.Hour * 24).Seconds()), "/", config.Conf.Domain, config.Conf.Secure, true)
	// Return empty string

	c.IndentedJSON(http.StatusOK, "")
}

func (f *PublicController) Register(c *gin.Context) {
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

	type User struct {
		Firstname           string `json:"firstname"`
		Surname             string `json:"surname"`
		Email               string `json:"email"`
		Password            string `json:"password"`
		RoleID              int    `json:"role_id"`
		PreferredLanguageID int    `json:"preferred_language_id"`
	}

	var tmpUser User
	if err := c.BindJSON(&tmpUser); err != nil {
		log.Errorf("Unable to bind json: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	pw := []byte(tmpUser.Password)
	newUser := models.User{
		Firstname:           tmpUser.Firstname,
		Surname:             tmpUser.Surname,
		Email:               tmpUser.Email,
		Password:            pw,
		RoleID:              tmpUser.RoleID,
		PreferredLanguageID: tmpUser.PreferredLanguageID,
	}

	id, err := dbi.CreateUser(f.Database, newUser)
	if err != nil {
		log.Errorf("Unable to create user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	newUser.ID = id
	newUser.Password = nil
	c.IndentedJSON(http.StatusCreated, newUser)
}

func (f *PublicController) UploadMaterial(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	if c.ContentType() == "text/plain" {
		type _file struct {
			Name string `json:"name"`
			Uri  string `json:"uri"`
		}

		var file _file
		if err := c.BindJSON(&file); err != nil {
			log.Errorf("Unable to bind json: %s\n", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
		log.Info(file)

		user_id, err := f.GetIdFromCookie(c)
		if err != nil {
			log.Errorf("Unable to get id from Cookie: %s\n", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
		coursematerial.CreateMaterial(f.Database, file.Name, file.Uri, user_id, id, false, nil)
	} else {
		file, err := c.FormFile("file")
		if err != nil {
			log.Errorf("No file found in request: %s", err.Error())
			c.Status(http.StatusBadRequest)
			return
		}

		fi, err := file.Open()
		if err != nil {
			log.Errorf("Unable to open file: %s", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}

		user_id, err := f.GetIdFromCookie(c)
		if err != nil {
			log.Errorf("Unable to get id from Cookie: %s\n", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
		err = coursematerial.CreateMaterial(f.Database, file.Filename, "", user_id, id, true, fi)
		if err != nil {
			log.Errorf("Unable to create CourseMaterial: %s", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.Status(http.StatusCreated)
}

func (f *PublicController) GetMaterialsFromCourse(c *gin.Context) {
	type _file struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		URI  string `json:"uri"`
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	files, err := coursematerial.GetAllMaterialsFromCourse(f.Database, id)
	if err != nil {
		log.Errorf("Unable to get all materials from course: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	var _files []_file
	for _, file := range files {
		uri := ""
		if file.Local == 0 {
			uri = file.URI
		}

		_files = append(_files, _file{file.ID, file.Name, uri})
	}

	c.IndentedJSON(http.StatusOK, _files)
}

func (f *PublicController) GetMaterialFromCourse(c *gin.Context) {
	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	file_id, err := strconv.Atoi(c.Param("file_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `file_id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	file, err := coursematerial.GetMaterialFromCourse(f.Database, course_id, file_id)
	if err != nil {
		log.Errorf("Unable to get material with id %d from course: %s", file_id, err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.File(file.URI)
	c.Status(http.StatusOK)
}

func (f *PublicController) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Error("Unable to convert parameter 'id' to an integer")
		return
	}

	_, err = dbi.GetUserById(f.Database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Errorf("User with id %d doesn't exist: %s", id, err.Error())
			c.Status(http.StatusNotFound)
			return
		}

		log.Errorf("Unable to get user with id %d", id)
		c.Status(http.StatusInternalServerError)
		return
	}

	err = dbi.DeleteUser(f.Database, id)
	if err != nil {
		log.Errorf("Unable to delete user from db: %s", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetUserById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		log.Error("Unable to convert parameter 'user_id' to an integer")
		c.Status(http.StatusInternalServerError)
		return
	}

	user, err := dbi.GetUserById(f.Database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Errorf("User with id %d doesn't exist: %s", id, err.Error())
			c.Status(http.StatusNotFound)
			return
		}

		log.Errorf("Unable to get user with id %d", id)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}

func (f *PublicController) GetAllAppointments(c *gin.Context) {

	/*
		user_id, err := f.GetIdFromCookie(c)
		if err != nil {
			log.Errorf("Unable to get id from Cookie: %s\n", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
	*/

	pCon := &calender.PublicController{Database: f.Database}
	appointments, err := pCon.GetAllAppointments(9999) // for testing, use 9999 instead of user_id
	if err != nil {
		log.Errorf("Unable to get all appointments from user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, appointments)
}

func (f *PublicController) GetAppointments(c *gin.Context) {

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
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
	startDate, err := time.Parse("2006-01-02", j["startDate"].(string))
	if err != nil {
		log.Error("unable to convert startDate to time.Time")
		c.Status(http.StatusInternalServerError)
		return
	}
	endDate, err := time.Parse("2006-01-02", j["endDate"].(string))
	if err != nil {
		log.Error("unable to convert endDate to time.Time")
		c.Status(http.StatusInternalServerError)
		return
	}

	pCon := &calender.PublicController{Database: f.Database}
	appointments, err := pCon.GetAppointments(user_id, startDate, endDate) // for testing, use 9999 instead of user_id
	if err != nil {
		log.Errorf("Unable to get appointments from user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, appointments)
}

func (f *PublicController) GetAllSubmissions(c *gin.Context) {

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	pCon := &calender.PublicController{Database: f.Database}
	appointments, err := pCon.GetAllSubmissions(user_id) // for testing, use 9999 instead of user_id
	if err != nil {
		log.Errorf("Unable to get submissions from user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, appointments)
}

func (f *PublicController) AddCourseToCalender(c *gin.Context) {

	var j map[string]interface{}

	if err := c.BindJSON(&j); err != nil {
		log.Errorf("Unable to bind json: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	dateStr, ok := j["date"].(string)
	if !ok {
		log.Error("unable to convert date to string")
		c.Status(http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Error("unable to convert string to time.Time")
		c.Status(http.StatusBadRequest)
		return
	}
	location, ok := j["location"].(string)
	if !ok {
		log.Error("unable to convert location to string")
		c.Status(http.StatusBadRequest)
		return
	}
	onlineStr, ok := j["online"].(string)
	if !ok {
		log.Error("unable to convert online to string")
		c.Status(http.StatusBadRequest)
		return
	}
	online, err := strconv.ParseInt(onlineStr, 10, 8)
	if err != nil {
		log.Error("unable to convert string to int8")
		c.Status(http.StatusBadRequest)
		return
	}
	courseIdStr, ok := j["courseId"].(string)
	if !ok {
		log.Error("unable to convert courseId to int")
		c.Status(http.StatusBadRequest)
		return
	}
	courseId, err := strconv.ParseInt(courseIdStr, 10, 64)
	if err != nil {
		log.Error("unable to convert string to int64")
		c.Status(http.StatusBadRequest)
		return
	}
	repeatsStr, ok := j["repeats"].(string)
	if !ok {
		log.Error("unable to convert repeats to bool")
		c.Status(http.StatusBadRequest)
		return
	}
	repeats, err := strconv.ParseBool(repeatsStr)
	if err != nil {
		log.Error("unable to convert string to bool")
		c.Status(http.StatusBadRequest)
		return
	}
	repeatDistanceStr, ok := j["repeatDistance"].(string)
	if !ok {
		log.Error("unable to convert repeatDistance to int")
		c.Status(http.StatusBadRequest)
		return
	}
	repeatDistance, err := strconv.ParseInt(repeatDistanceStr, 10, 64)
	if err != nil {
		log.Error("unable to convert string to int64")
		c.Status(http.StatusBadRequest)
		return
	}
	repeatEnd, err := time.Parse("2006-01-02", j["repeatEnd"].(string))
	if err != nil {
		log.Error("unable to convert repeatEnd to string")
		c.Status(http.StatusBadRequest)
		return
	}

	pCon := &calender.PublicController{Database: f.Database}
	_, err = pCon.AddCourseToCalender(date, null.StringFrom(location), int8(online), int(courseId), repeats, int(repeatDistance), repeatEnd)
	if err != nil {
		log.Errorf("Unable to create course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.Status(http.StatusOK)

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

	submDate, err := time.Parse("2006-01-02", j["submDate"].(string))
	if err != nil {
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

	pCon := &calender.PublicController{Database: f.Database}
	id, err := pCon.AddSubmissionToCalender(submDate, submName, courseId)
	if err != nil {
		log.Errorf("Unable to add appointment of submission: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	newAppointment.ID = id
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
	pCon := &calender.PublicController{Database: f.Database}
	err = pCon.DeactivateAppointment(appointment_id)
	if err != nil {
		log.Errorf("Unable to delete appointment: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
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
	pCon := &calender.PublicController{Database: f.Database}
	err = pCon.DeactivateCourseInCalender(appointment_id, course_id, repeats)
	if err != nil {
		log.Errorf("Unable to delete appointment from course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
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
	pCon := &calender.PublicController{Database: f.Database}
	err = pCon.DeactivateExamInCalender(appointment_id, exam_id)
	if err != nil {
		log.Errorf("Unable to delete exam from calender: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Status(http.StatusNoContent)

}

func (f *PublicController) SearchCourse(c *gin.Context) {

	searchterm, ok := c.GetQuery("searchterm")
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, errors.New("query searchterm not found"))
		return
	}

	courses, err := course.SearchCourse(f.Database, searchterm)

	if err != nil {
		log.Errorf("Unable to search course: %s\n", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, courses)
}
