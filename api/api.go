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
	"learningbay24.de/backend/institution"
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
	//Return Status and Data in JSON-Format

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
	//Return Status and Data in JSON-Format

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
	//Return Status and Data in JSON-Format

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
	tmp, ok = j["role_id"].(float64)
	if !ok {
		log.Error("unable to convert user_id to float64")
		c.Status(http.StatusInternalServerError)
		return
	}
	role_id := int(tmp)

	id, err := course.CreateCourse(f.Database, name, null.StringFrom(description), enroll_key, user_id, role_id)
	if err != nil {
		log.Errorf("Unable to create course: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, id)
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

func (f *PublicController) EditCourseById(c *gin.Context) {

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
	_, err = course.EditCourse(f.Database, id, newCourse.Name, newCourse.Description, newCourse.EnrollKey)
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
	c.SetCookie("user_token", token, int((time.Hour * 24).Seconds()), "/", config.Conf.Domain, config.Conf.Secure, true)
	// Return user with set cookie
	newUser.Password = nil
	newUser.ID = id
	c.IndentedJSON(http.StatusOK, newUser)

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

/* Uncomment, when calender.go is integrated into main branch

func (f *PublicController) GetAllAppointments(c *gin.Context) {

	user_id, err := strconv.Atoi(c.Param("user_id"))

		appointments, err := calender.GetAllAppointments(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get appointments from user: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, appointments)
}
*/
func (f *PublicController) GetUserCount(c *gin.Context) {

	//Fetch Data from Database with Backend function
	usercount, err := institution.GetUserCount(f.Database)
	if err != nil {
		log.Errorf("Unable to get usercount: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, usercount)

}

func (f *PublicController) GetAllFieldsOfStudy(c *gin.Context) {

	//Fetch Data from Database with Backend function
	fieldofstudies, err := institution.GetAllFieldsOfStudies(f.Database)
	if err != nil {
		log.Errorf("Unable to get all FieldOfStudies: %s\n", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	//Return Status and Data in JSON-Format
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, fieldofstudies)

}

func (f *PublicController) CreateFieldOfStudy(c *gin.Context) {

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

	tmp, ok := j["semesters"].(float64)
	if !ok {
		log.Error("unable to convert semesters to float64")
		c.Status(http.StatusBadRequest)
		return
	}
	semesters := int(tmp)

	id, err := institution.CreateFieldOfStudy(f.Database, null.StringFrom(name), null.IntFrom(semesters))
	if err != nil {
		log.Errorf("Unable to create fieldofstudy: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, id)
}

func (f *PublicController) DeleteFieldOfStudy(c *gin.Context) {

	//Get given ID from the Context
	//Convert data type from str to int to use ist as param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	//Deactivate Data from Database with Backend function
	fos_id, err := institution.DeleteFieldOfStudy(f.Database, id)
	//Return Status and Data in JSON-Format
	if err != nil {
		log.Errorf("Unable to delete fieldofstudy: %s\n", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, fos_id)
}

func (f *PublicController) EditFieldOfStudyById(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var newFieldOfStudy models.FieldOfStudy
	if err := c.BindJSON(&newFieldOfStudy); err != nil {
		log.Errorf("Unable to bind json: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	_, err = institution.EditFieldOfStudy(f.Database, id, newFieldOfStudy.Name.String, newFieldOfStudy.Semesters.Int)
	if err != nil {
		log.Errorf("Unable to update fieldofstudy: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, id)
}

func (f *PublicController) AddFieldOfStudyHasCourse(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to string: %s\n", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	fieldOfStudy_id, err := strconv.Atoi(c.Param("fieldofstudy_id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var newfos_has_course models.FieldOfStudyHasCourse
	if err := c.BindJSON(&newfos_has_course); err != nil {
		if err != nil {
			log.Errorf("Unable to bind json: %s\n", err.Error())
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}
	}

	err = institution.AddFieldOfStudyHasCourse(f.Database, fieldOfStudy_id, id, newfos_has_course.Semester)
	if err != nil {
		log.Errorf("Unable to add fieldOfStudy to course: %s\n", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, newfos_has_course)
}

func (f *PublicController) DeleteFieldOfStudyHasCourse(c *gin.Context) {

	//Get given ID from the Context
	//Convert data type from str to int to use ist as param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	fieldOfStudy_id, err := strconv.Atoi(c.Param("fieldofstudy_id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	//Deactivate Data from Database with Backend function
	err = institution.DeleteFieldOfStudyHasCourse(f.Database, fieldOfStudy_id, id)
	//Return Status and Data in JSON-Format
	if err != nil {
		log.Errorf("Unable to delete fieldofstudy from course: %s\n", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, fieldOfStudy_id)
}

func (f *PublicController) EditFieldOfStudyHasCourseById(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	fieldOfStudy_id, err := strconv.Atoi(c.Param("fieldofstudy_id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var newFieldOfStudyHasCourse models.FieldOfStudyHasCourse
	if err := c.BindJSON(&newFieldOfStudyHasCourse); err != nil {
		log.Errorf("Unable to bind json: %s\n", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	err = institution.EditFieldOfStudyHasCourse(f.Database, fieldOfStudy_id, id, newFieldOfStudyHasCourse.Semester)
	if err != nil {
		log.Errorf("Unable to update fieldofstudy: %s\n", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	newFieldOfStudyHasCourse.CourseID = id
	newFieldOfStudyHasCourse.FieldOfStudyID = fieldOfStudy_id

	c.IndentedJSON(http.StatusOK, newFieldOfStudyHasCourse)
}
