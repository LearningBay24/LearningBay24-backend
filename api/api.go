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

	"learningbay24.de/backend/config"
	"learningbay24.de/backend/course"
	coursematerial "learningbay24.de/backend/courseMaterial"
	"learningbay24.de/backend/dbi"
	"learningbay24.de/backend/exam"
	"learningbay24.de/backend/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
)

type PublicController struct {
	Database *sql.DB
}

func AuthorizeAdmin(roleId int) bool {
	log.Infof("Authorizing with role id: %d", roleId)
	return roleId <= dbi.AdminRoleId
}

func AuthorizeModerator(roleId int) bool {
	log.Infof("Authorizing with role id: %d", roleId)
	return roleId <= dbi.ModeratorRoleId
}

func AuthorizeUser(roleId int) bool {
	log.Infof("Authorizing with role id: %d", roleId)
	return roleId <= dbi.UserRoleId
}

func AuthorizeCourseAdmin(roleId int) bool {
	log.Infof("Authorizing with role id: %d", roleId)
	return roleId <= dbi.CourseAdminRoleId
}

func AuthorizeCourseModerator(roleId int) bool {
	log.Infof("Authorizing with role id: %d", roleId)
	return roleId <= dbi.CourseModeratorRoleId
}

func AuthorizeCourseUser(roleId int) bool {
	log.Infof("Authorizing with role id: %d", roleId)
	return roleId <= dbi.CourseUserRoleId
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
		log.Errorf("Error parsing token: %s", err.Error())
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
		log.Errorf("Unable to get Data from Cookie: %s", err.Error())
		return 0, err
	}
	datamap, ok := data.(map[string]interface{})
	if !ok {
		return 0, errors.New("Unable to map id from data interface")
	}
	id, err := strconv.Atoi(datamap["id"].(string))
	if err != nil {
		log.Errorf("Unable to convert idstring to int: %s", err.Error())
		return 0, err
	}
	return id, err

}

func (f *PublicController) GetRoleIdFromCookie(c *gin.Context) (int, error) {
	data, err := f.GetDataFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get Data from Cookie: %s", err.Error())
		return 0, err
	}
	datamap, ok := data.(map[string]interface{})
	if !ok {
		return 0, errors.New("Unable to map id from data interface")
	}
	id, err := strconv.Atoi(datamap["role_id"].(string))
	if err != nil {
		log.Errorf("Unable to convert idstring to int: %s", err.Error())
		return 0, err
	}
	return id, err
}

func (f *PublicController) GetCourseById(c *gin.Context) {
	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !AuthorizeCourseUser(course_role) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	// Fetch Data from Database with Backend function
	course, err := course.GetCourse(f.Database, course_id)
	if err != nil {
		log.Errorf("Unable to get course: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, course)
}

func (f *PublicController) DeleteUserFromCourse(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !AuthorizeCourseAdmin(course_role) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	user_to_delete_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `user_id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	// Fetch Data from Database with Backend function
	err = course.DeleteUserFromCourse(f.Database, user_to_delete_id, course_id)
	if err != nil {
		log.Errorf("Unable to delete user from course: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetUsersInCourse(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !AuthorizeCourseUser(course_role) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}
	// Fetch Data from Database with Backend function
	users, err := course.GetUsersInCourse(f.Database, course_id)
	if err != nil {
		log.Errorf("Unable to get users in course: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, users)
}

func (f *PublicController) GetCoursesFromUser(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	role_id, err := f.GetRoleIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get role_id from cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	if !AuthorizeUser(role_id) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Fetch Data from Database with Backend function
	courses, err := course.GetCoursesFromUser(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get courses from user: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, courses)
}

func (f *PublicController) DeleteCourse(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !AuthorizeCourseAdmin(course_role) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}
	// Deactivate Data from Database with Backend function
	course, err := course.DeleteCourse(f.Database, course_id)
	// Return Status and Data in JSON-Format
	if err != nil {
		log.Errorf("Unable to delete course: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, course)
}

func (f *PublicController) CreateCourse(c *gin.Context) {
	role_id, err := f.GetRoleIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get role_id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusUnauthorized, err.Error())
		return
	}

	if !AuthorizeModerator(role_id) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	var newCourse models.Course

	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
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
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
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
		log.Errorf("Unable to create course: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	newCourse.ID = id

	c.IndentedJSON(http.StatusOK, newCourse)
}

func (f *PublicController) EnrollUser(c *gin.Context) {
	role_id, err := f.GetRoleIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get role_id from cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	if !AuthorizeUser(role_id) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		if err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
	}

	_, err = course.EnrollUser(f.Database, user_id, id, newCourse.EnrollKey)
	if err != nil {
		log.Errorf("Unable to enroll user in course: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, newCourse)
}

func (f *PublicController) UpdateCourseById(c *gin.Context) {

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !AuthorizeCourseModerator(course_role) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}
	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		log.Errorf("Unable to bind json: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	_, err = course.UpdateCourse(f.Database, course_id, newCourse.Name, newCourse.Description, newCourse.EnrollKey)
	if err != nil {
		log.Errorf("Unable to update course: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	newCourse.ID = course_id
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
		log.Errorf("Unable to bind json: %s", err.Error())
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
			log.Errorf("Unable to verify credentials: %s", err.Error())
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
		log.Errorf("Unable to sign token: %s", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Set the cookie and add it to the response header
	c.SetCookie("user_token", tokenString, int((time.Hour * 24).Seconds()), "/", config.Conf.Domain, config.Conf.Secure, true)
	// Return empty string

	c.IndentedJSON(http.StatusOK, "")
}

func (f *PublicController) Logout(c *gin.Context) {

	role_id, err := f.GetRoleIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get role_id from cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	if !AuthorizeUser(role_id) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	c.SetCookie("user_token", "", -1, "/", config.Conf.Domain, config.Conf.Secure, true)
	c.IndentedJSON(http.StatusOK, "")
}

func (f *PublicController) Register(c *gin.Context) {
	role_id, err := f.GetRoleIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get role_id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusUnauthorized, err.Error())
		return
	}

	if !AuthorizeModerator(role_id) {
		log.Infof("User is not authorized: %s", err.Error())
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
		log.Errorf("Unable to bind json: %s", err.Error())
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
		log.Errorf("Unable to create user: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	newUser.ID = id
	newUser.Password = nil
	c.IndentedJSON(http.StatusCreated, newUser)
}

func (f *PublicController) UploadMaterial(c *gin.Context) {

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !AuthorizeCourseModerator(course_role) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	if c.ContentType() == "text/plain" {
		type _file struct {
			Name string `json:"name"`
			Uri  string `json:"uri"`
		}

		var file _file
		if err := c.BindJSON(&file); err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
		log.Info(file)

		user_id, err := f.GetIdFromCookie(c)
		if err != nil {
			log.Errorf("Unable to get id from Cookie: %s", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
		err = coursematerial.CreateMaterial(f.Database, file.Name, file.Uri, user_id, course_id, false, nil)
		if err != nil {
			log.Errorf("Unable to create CourseMaterial: %s", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
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
			log.Errorf("Unable to get id from Cookie: %s", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
		err = coursematerial.CreateMaterial(f.Database, file.Filename, "", user_id, course_id, true, fi)
		if err != nil {
			log.Errorf("Unable to create CourseMaterial: %s", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.Status(http.StatusCreated)
}

func (f *PublicController) GetMaterialsFromCourse(c *gin.Context) {
	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !AuthorizeCourseUser(course_role) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	type _file struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		URI  string `json:"uri"`
	}

	files, err := coursematerial.GetAllMaterialsFromCourse(f.Database, course_id)
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

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !AuthorizeCourseUser(course_role) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
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

	role_id, err := f.GetRoleIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get role_id from cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	if !AuthorizeAdmin(role_id) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
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

func (f *PublicController) GetUserByCookie(c *gin.Context) {

	role_id, err := f.GetRoleIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get role_id from cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	if !AuthorizeUser(role_id) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := dbi.GetUserById(f.Database, user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Errorf("User with id %d doesn't exist: %s", user_id, err.Error())
			c.Status(http.StatusNotFound)
			return
		}

		log.Errorf("Unable to get user with id %d", user_id)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}

func (f *PublicController) GetUserById(c *gin.Context) {

	role_id, err := f.GetRoleIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get role_id from cookie: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	if !AuthorizeUser(role_id) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	user_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := dbi.GetUserById(f.Database, user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Errorf("User with id %d doesn't exist: %s", user_id, err.Error())
			c.Status(http.StatusNotFound)
			return
		}

		log.Errorf("Unable to get user with id %d", user_id)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}

/* Uncomment, when calender.go is integrated into main branch

func (f *PublicController) GetAllAppointments(c *gin.Context) {

	user_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	appointments, err := calender.GetAllAppointments(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get appointments from user: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, appointments)
}
*/

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

func (f *PublicController) CreateExam(c *gin.Context) {
	var newExam models.Exam

	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
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
		c.Status(http.StatusBadRequest)
		return
	}

	description, ok := j["description"].(string)
	if !ok {
		log.Error("unable to convert description to string")
		c.Status(http.StatusBadRequest)
		return
	}

	dateStr, ok := j["date"].(string)
	if !ok {
		log.Error("unable to convert date to string")
		c.Status(http.StatusBadRequest)
		return
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `date` to time.Time: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	durationStr, ok := j["duration"].(string)
	if !ok {
		log.Error("unable to convert duration to string")
		c.Status(http.StatusBadRequest)
		return
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `duration` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	courseIdStr, ok := j["course_id"].(string)
	if !ok {
		log.Error("unable to convert course_id to string")
		c.Status(http.StatusBadRequest)
		return
	}

	courseId, err := strconv.Atoi(courseIdStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `course_id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	creatorId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	course_role, err := course.GetCourseRole(f.Database, creatorId, courseId)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !AuthorizeCourseAdmin(course_role) {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}

	onlineStr, ok := j["online"].(string)
	if !ok {
		log.Error("unable to convert online to string")
		c.Status(http.StatusBadRequest)
		return
	}

	online, err := strconv.Atoi(onlineStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `online` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
	}

	location, ok := j["location"].(string)
	if !ok {
		log.Error("unable to convert location to string")
		c.Status(http.StatusBadRequest)
		return
	}

	registerDeadlineStr, ok := j["register_deadline"].(string)
	if !ok {
		log.Error("unable to convert register_deadline to string")
		c.Status(http.StatusBadRequest)
		return
	}

	registerDeadline, err := time.Parse(time.RFC3339, registerDeadlineStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `register_deadline` to time.Time: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	deregisterDeadlineStr, ok := j["deregister_deadline"].(string)
	if !ok {
		log.Error("unable to convert deregister_deadline to string")
		c.Status(http.StatusBadRequest)
		return
	}

	deregisterDeadline, err := time.Parse(time.RFC3339, deregisterDeadlineStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `deregister_deadline` to time.Time: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	id, err := pCtrl.CreateExam(name, description, date, duration, courseId, creatorId, int8(online), null.StringFrom(location), null.TimeFrom(registerDeadline), null.TimeFrom(deregisterDeadline))
	if err != nil {
		log.Errorf("Unable to create exam: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	newExam.ID = id

	c.IndentedJSON(http.StatusCreated, newExam)
}

func (f *PublicController) EditExam(c *gin.Context) {
	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Errorf("Unable to unmarshal the json body: %+v", raw)
		c.Status(http.StatusBadRequest)
		return
	}

	name, ok := j["name"].(string)
	if !ok {
		log.Error("unable to convert name to string")
		c.Status(http.StatusBadRequest)
		return
	}

	description, ok := j["description"].(string)
	if !ok {
		log.Error("unable to convert description to string")
		c.Status(http.StatusBadRequest)
		return
	}

	dateStr, ok := j["date"].(string)
	if !ok {
		log.Error("unable to convert date to string")
		c.Status(http.StatusBadRequest)
		return
	}
	var date time.Time
	if dateStr != "" {
		date, err = time.ParseInLocation(time.RFC3339, dateStr, time.Local)
		if err != nil {
			log.Errorf("Unable to convert parameter `date` to time.Time: %s", err.Error())
			c.Status(http.StatusBadRequest)
			return
		}
		date = date.Local()
	}
	durationStr, ok := j["duration"].(string)
	if !ok {
		log.Error("unable to convert duration to string")
		c.Status(http.StatusBadRequest)
		return
	}
	var duration int
	if durationStr != "" {
		duration, err = strconv.Atoi(durationStr)
		if err != nil {
			log.Errorf("Unable to convert parameter `duration` to int: %s", err.Error())
			c.Status(http.StatusBadRequest)
			return
		}
	}

	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	userId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	onlineStr, ok := j["online"].(string)
	if !ok {
		log.Error("unable to convert online to string")
		c.Status(http.StatusBadRequest)
		return
	}
	var online null.Int8
	if onlineStr != "" {
		onlineInt, err := strconv.Atoi(onlineStr)
		if err != nil {
			log.Errorf("Unable to convert parameter `online` to int: %s", err.Error())
			c.Status(http.StatusBadRequest)
		}
		online.Int8 = int8(onlineInt)
		online.Valid = true
	} else {
		online.Valid = false
	}

	location, ok := j["location"].(string)
	if !ok {
		log.Error("unable to convert location to string")
		c.Status(http.StatusBadRequest)
		return
	}

	registerDeadlineStr, ok := j["register_deadline"].(string)
	if !ok {
		log.Error("unable to convert register_deadline to string")
		c.Status(http.StatusBadRequest)
		return
	}
	var registerDeadline time.Time
	if registerDeadlineStr != "" {
		registerDeadline, err = time.ParseInLocation(time.RFC3339, registerDeadlineStr, time.Local)
		if err != nil {
			log.Errorf("Unable to convert parameter `register_deadline` to time.Time: %s", err.Error())
			c.Status(http.StatusBadRequest)
			return
		}
		registerDeadline = registerDeadline.Local()
	}

	deregisterDeadlineStr, ok := j["deregister_deadline"].(string)
	if !ok {
		log.Error("unable to convert register_deadline to string")
		c.Status(http.StatusBadRequest)
		return
	}
	var deregisterDeadline time.Time
	if deregisterDeadlineStr != "" {
		deregisterDeadline, err = time.ParseInLocation(time.RFC3339, deregisterDeadlineStr, time.Local)
		if err != nil {
			log.Errorf("Unable to convert parameter `deregister_deadline` to time.Time: %s", err.Error())
			c.Status(http.StatusBadRequest)
			return
		}
		deregisterDeadline = deregisterDeadline.Local()
	}

	pCtrl := exam.PublicController{Database: f.Database}
	err = pCtrl.EditExam(name, description, date, duration, examId, userId, online, null.StringFrom(location), null.TimeFrom(registerDeadline), null.TimeFrom(deregisterDeadline))
	if err != nil {
		log.Errorf("Unable to edit exam: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (f *PublicController) UploadExamFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())

		return
	}

	if c.ContentType() == "text/plain" {
		type _file struct {
			Name string `json:"name"`
			Uri  string `json:"uri"`
		}

		var file _file
		if err := c.BindJSON(&file); err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
		log.Info(file)

		pCtrl := exam.PublicController{Database: f.Database}
		err = pCtrl.UploadExamFile(file.Name, file.Uri, user_id, id, false, nil)
		if err != nil {
			log.Errorf("Unable to create Exam-URI: %s", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
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

		pCtrl := exam.PublicController{Database: f.Database}
		err = pCtrl.UploadExamFile(file.Filename, "", user_id, id, true, fi)
		if err != nil {
			log.Errorf("Unable to create ExamFile: %s", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.Status(http.StatusCreated)
}

func (f *PublicController) GetExamById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	// Fetch Data from Database with Backend function
	pCtrl := exam.PublicController{Database: f.Database}
	co, err := pCtrl.GetExamByID(id)
	if err != nil {
		log.Errorf("Unable to get exam: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, co)
}

func (f *PublicController) GetRegisteredExamsFromUser(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	userId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Fetch Data from Database with Backend function
	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetRegisteredExamsFromUser(userId)
	if err != nil {
		log.Errorf("Unable to get exams from user: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetUnregisteredExamsFromUser(c *gin.Context) {
	userId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetUnregisteredExams(userId)
	if err != nil {
		log.Errorf("Unable to get exams: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetExamsFromCourse(c *gin.Context) {
	courseId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	// Fetch Data from Database with Backend function
	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetExamsFromCourse(courseId)
	if err != nil {
		log.Errorf("Unable to get exam: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetAttendedExamsFromUser(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Fetch Data from Database with Backend function
	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetAttendedExamsFromUser(user_id)
	if err != nil {
		log.Errorf("Unable to get exams from user: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetPassedExamsFromUser(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Fetch Data from Database with Backend function
	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetPassedExamsFromUser(user_id)
	if err != nil {
		log.Errorf("Unable to get exams from user: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetCreatedFromUser(c *gin.Context) {
	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	user_id, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Fetch Data from Database with Backend function
	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetCreatedExamsFromUser(user_id)
	if err != nil {
		log.Errorf("Unable to get exams from exam: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) RegisterToExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'exam_id' to an int: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	userId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get user_id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	user, err := pCtrl.RegisterToExam(userId, examId)
	if err != nil {
		log.Errorf("Unable to register user to course: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}

func (f *PublicController) DeregisterFromExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'exam_id' to an int: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	userId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get user_id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	err = pCtrl.DeregisterFromExam(userId, examId)
	if err != nil {
		log.Errorf("Unable to deregister user from exam: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetFileFromExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	userId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get user_id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	file, err := pCtrl.GetFileFromExam(examId)
	if err != nil {
		log.Errorf("Unable to get file from exam: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	err = pCtrl.AttendExam(examId, userId)
	if err != nil {
		log.Errorf("Unable to get file from exam: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.File(file[0].URI)
	c.Status(http.StatusOK)
}

func (f *PublicController) SubmitAnswerToExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	userId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if c.ContentType() == "text/plain" {
		type _file struct {
			Name string `json:"name"`
			Uri  string `json:"uri"`
		}

		var file _file
		if err := c.BindJSON(&file); err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
		log.Info(file)

		pCtrl := exam.PublicController{Database: f.Database}
		err = pCtrl.SubmitAnswer(file.Name, file.Uri, examId, userId, false, nil)
		if err != nil {
			log.Errorf("Unable to submit answer: %s", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
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
		pCtrl := exam.PublicController{Database: f.Database}
		err = pCtrl.SubmitAnswer(file.Filename, "", examId, userId, true, fi)
		if err != nil {
			log.Errorf("Unable to submit answer: %s", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	pCtrl := exam.PublicController{Database: f.Database}
	err = pCtrl.AttendExam(examId, userId)
	if err != nil {
		log.Errorf("Unable to attend user to exam: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.Status(http.StatusCreated)
}

func (f *PublicController) GetRegisteredUsersFromExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'id' to an int: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	creatorId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	attendees, err := pCtrl.GetRegisteredUsersFromExam(examId, creatorId)
	if err != nil {
		log.Errorf("Unable to fetch attendees from exam: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, attendees)
}

func (f *PublicController) GetAttendeesFromExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'id' to an int: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	creatorId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	attendees, err := pCtrl.GetAttendeesFromExam(examId, creatorId)
	if err != nil {
		log.Errorf("Unable to fetch attendees from exam: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, attendees)
}

func (f *PublicController) GetFileFromAttendee(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	examId, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'exam_id' to an int: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	file, err := pCtrl.GetAnswerFromAttendee(userId, examId)
	if err != nil {
		log.Errorf("Unable to get file from answer: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.File(file.URI)
	c.Status(http.StatusOK)
}

func (f *PublicController) GradeAnswer(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `exam_id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	creatorId, err := f.GetIdFromCookie(c)
	if err != nil {
		log.Errorf("Unable to get id from Cookie: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `user_id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
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

	gradeStr, ok := j["grade"].(string)
	if !ok {
		log.Error("unable to convert grade to string")
		c.Status(http.StatusInternalServerError)
		return
	}

	grade, err := strconv.Atoi(gradeStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `grade` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	passedStr, ok := j["passed"].(string)
	if !ok {
		log.Error("unable to convert passed to string")
		c.Status(http.StatusInternalServerError)
		return
	}

	passed, err := strconv.Atoi(passedStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `passed` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	feedback, ok := j["feedback"].(string)
	if !ok {
		log.Error("unable to convert feedback to string")
		c.Status(http.StatusInternalServerError)
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	err = pCtrl.GradeAnswer(examId, creatorId, userId, null.IntFrom(grade), null.Int8From(int8(passed)), null.StringFrom(feedback))
	if err != nil {
		log.Errorf("Unable to grade answer: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (f *PublicController) SetAttended(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `exam_id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `user_id` to int: %s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	err = pCtrl.SetAttended(examId, userId)
	if err != nil {
		log.Errorf("Unable to set user's exam to attended: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (f *PublicController) DeleteExam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	// Deactivate Data from Database with Backend function
	pCtrl := exam.PublicController{Database: f.Database}
	ex, err := pCtrl.DeleteExam(id)
	// Return Status and Data in JSON-Format
	if err != nil {
		log.Errorf("Unable to delete course: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, ex)
}
