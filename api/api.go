package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"learningbay24.de/backend/calender"
	"learningbay24.de/backend/config"
	"learningbay24.de/backend/course"
	coursematerial "learningbay24.de/backend/courseMaterial"
	"learningbay24.de/backend/dbi"
	"learningbay24.de/backend/errs"
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

type _file struct {
	Name string `json:"name"`
	Uri  string `json:"uri"`
}

// Handle an error by setting the correct HTTP status code and filling the body of the response with the error message if necessary.
func handleApiError(c *gin.Context, err error) {
	NOT_AUTHORIZED := []error{errs.ErrNotAdmin, errs.ErrNotModerator, errs.ErrNotUser, errs.ErrNotCourseAdmin, errs.ErrNotCourseModerator, errs.ErrNotCourseUser}
	NOT_FOUNDS := []error{sql.ErrNoRows, errs.ErrNoUploads}
	BAD_REQUESTS := []error{errs.ErrFileExtensionNotAllowed, errs.ErrNoFileExtension, errs.ErrParameterConversion, errs.ErrNoFileInRequest, errs.ErrBodyConversion, errs.ErrNoQuery, errs.ErrRawData, errs.ErrUploadLimitReached, errs.ErrEmptyName, errs.ErrVisibleTimePast, errs.ErrDeadlineTimePast, errs.ErrVisibleFromAfterDeadline, errs.ErrSubmissionTimeAfterDeadline, errs.ErrEmptyFileName}
	CONFLICTS := []error{errs.ErrSelfRegisterExam, errs.ErrRegisterDeadlinePassed, errs.ErrUnregisterDeadlinePassed, errs.ErrExamEnded, errs.ErrExamHasntStarted, errs.ErrCourseNotEmpty, errs.ErrWrongEnrollkey}

	log.Error(err)

	for _, na := range NOT_AUTHORIZED {
		if errors.Is(err, na) {
			c.Status(http.StatusForbidden)
			return
		}
	}

	for _, nf := range NOT_FOUNDS {
		if errors.Is(err, nf) {
			c.Status(http.StatusNotFound)
			return
		}
	}

	for _, br := range BAD_REQUESTS {
		if errors.Is(err, br) {
			c.JSON(http.StatusBadRequest, br)
			return
		}
	}

	for _, cf := range CONFLICTS {
		if errors.Is(err, cf) {
			c.Status(http.StatusConflict)
			return
		}
	}

	log.Debugf("Error not listed, type: %t", err)

	// internal server error if something else happened
	c.Status(http.StatusInternalServerError)
}

func AuthorizeAdmin(role_id int) bool {
	return role_id <= dbi.AdminRoleId
}

func AuthorizeModerator(role_id int) bool {
	return role_id <= dbi.ModeratorRoleId
}

func AuthorizeUser(role_id int) bool {
	return role_id <= dbi.UserRoleId
}

func AuthorizeCourseAdmin(course_role_id int, role_id int) bool {
	return course_role_id <= dbi.CourseAdminRoleId || AuthorizeAdmin(role_id)
}

func AuthorizeCourseModerator(course_role_id int, role_id int) bool {
	return course_role_id <= dbi.CourseModeratorRoleId || AuthorizeAdmin(role_id)
}

func AuthorizeCourseUser(course_role_id int, role_id int) bool {
	return course_role_id <= dbi.CourseUserRoleId || AuthorizeAdmin(role_id)
}

func (f *PublicController) AuthorizeUserHasExam(userId, examId int) (bool, error) {
	log.Infof("Authorizing exam id: %d with user id: %d", examId, userId)
	return models.UserHasExamExists(context.Background(), f.Database, userId, examId)
}

func (f *PublicController) GetCourseById(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	// Get given ID from the Context
	// Convert data type from str to int to use ist as param
	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}

	// Fetch Data from Database with Backend function
	course, err := course.GetCourse(f.Database, course_id)
	if err != nil {
		handleApiError(c, err)
		return
	}
	// Return Status and Data in JSON-Format
	c.IndentedJSON(http.StatusOK, course)
}

func (f *PublicController) DeleteUserFromCourse(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseAdmin(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseAdmin)
		return
	}

	user_to_delete_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `user_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	err = course.DeleteUserFromCourse(f.Database, user_to_delete_id, course_id)
	if err != nil {
		log.Errorf("Unable to delete user from course: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetUsersInCourse(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}

	users, err := course.GetUsersInCourse(f.Database, course_id)
	if err != nil {
		log.Errorf("Unable to get users in course: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, users)
}

func (f *PublicController) GetEnrolledCoursesFromUser(c *gin.Context) {
	role_id := c.MustGet("CookieRoleId").(int)

	if !AuthorizeUser(role_id) {
		handleApiError(c, errs.ErrNotUser)
		return
	}

	user_id := c.MustGet("CookieUserId").(int)

	courses, err := course.GetEnrolledCoursesFromUser(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get courses from user: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courses)
}

func (f *PublicController) GetCreatedCoursesFromUser(c *gin.Context) {
	role_id := c.MustGet("CookieRoleId").(int)

	if !AuthorizeUser(role_id) {
		handleApiError(c, errs.ErrNotUser)
		return
	}

	user_id := c.MustGet("CookieUserId").(int)

	courses, err := course.GetCreatedCoursesFromUser(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get courses from user: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courses)
}

func (f *PublicController) DeleteCourse(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseAdmin(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseAdmin)
		return
	}

	course, err := course.DeleteCourse(f.Database, course_id)
	if err != nil {
		log.Errorf("Unable to delete course: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, course)
}

func (f *PublicController) CreateCourse(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	if !AuthorizeModerator(role_id) {
		handleApiError(c, errs.ErrNotModerator)
		return
	}

	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		if err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			// NOTE: `BindJSON` sets the return status arleady
			return
		}
	}

	id, err := course.CreateCourse(f.Database, newCourse.Name, newCourse.Description, newCourse.EnrollKey, user_id)
	if err != nil {
		log.Errorf("Unable to create course: %s", err.Error())
		handleApiError(c, err)
		return
	}
	newCourse.ID = id

	c.IndentedJSON(http.StatusOK, newCourse)
}

func (f *PublicController) EnrollUser(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	if !AuthorizeUser(role_id) {
		handleApiError(c, errs.ErrNotUser)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		if err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			// NOTE: `BindJSON` sets the return status arleady
			return
		}
	}

	user, err := course.EnrollUser(f.Database, user_id, id, newCourse.EnrollKey)
	if err != nil {
		log.Errorf("Unable to enroll user in course: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, user.ID)
}

func (f *PublicController) EditCourseById(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}
	var newCourse models.Course
	if err := c.BindJSON(&newCourse); err != nil {
		log.Errorf("Unable to bind json: %s", err.Error())
		// NOTE: `BindJSON` sets the return status arleady
		return
	}
	_, err = course.EditCourse(f.Database, course_id, newCourse.Name, newCourse.Description, newCourse.EnrollKey)
	if err != nil {
		log.Errorf("Unable to update course: %s", err.Error())
		handleApiError(c, err)
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
		// NOTE: `BindJSON` sets the return status arleady
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
		handleApiError(c, err)
		return
	}

	user, err := dbi.GetUserById(f.Database, id)
	if err != nil {
		log.Errorf("Unable to get user by id: %s", err.Error())
		handleApiError(c, err)
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
		handleApiError(c, err)
		return
	}

	// Set the cookie and add it to the response header
	c.SetCookie("user_token", tokenString, int((time.Hour * 24).Seconds()), "/", config.Conf.Domain, config.Conf.Secure, true)

	c.Status(http.StatusOK)
}

func (f *PublicController) Logout(c *gin.Context) {
	role_id := c.MustGet("CookieRoleId").(int)

	if !AuthorizeUser(role_id) {
		handleApiError(c, errs.ErrNotUser)
		return
	}

	c.SetCookie("user_token", "", -1, "/", config.Conf.Domain, config.Conf.Secure, true)
	c.Status(http.StatusOK)
}

func (f *PublicController) Register(c *gin.Context) {
	role_id := c.MustGet("CookieRoleId").(int)

	if !AuthorizeModerator(role_id) {
		handleApiError(c, errs.ErrNotModerator)
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
		// NOTE: `BindJSON` sets the return status arleady
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
		handleApiError(c, err)
		return
	}

	newUser.ID = id
	newUser.Password = nil

	c.IndentedJSON(http.StatusCreated, newUser)
}

func (f *PublicController) UploadMaterial(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	if c.ContentType() == "text/plain" {
		var file _file
		if err := c.BindJSON(&file); err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			// NOTE: `BindJSON` sets the return status arleady
			return
		}

		if err := coursematerial.CreateMaterial(f.Database, file.Name, file.Uri, user_id, course_id, false, nil, 0); err != nil {
			log.Errorf("Unable to create CourseMaterial: %s", err.Error())
			handleApiError(c, err)
			return
		}
	} else {
		file, err := c.FormFile("file")
		if err != nil {
			log.Error(err)
			handleApiError(c, errs.ErrNoFileInRequest)
			return
		}

		fi, err := file.Open()
		if err != nil {
			handleApiError(c, err)
			return
		}

		err = coursematerial.CreateMaterial(f.Database, file.Filename, "", user_id, course_id, true, fi, int(file.Size))
		if err != nil {
			log.Errorf("Unable to create CourseMaterial: %s", err.Error())
			handleApiError(c, err)
			return
		}
	}

	c.Status(http.StatusCreated)
}

func (f *PublicController) GetMaterialsFromCourse(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
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
		handleApiError(c, err)
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
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}

	file_id, err := strconv.Atoi(c.Param("file_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `file_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	file, err := coursematerial.GetMaterialFromCourse(f.Database, course_id, file_id)
	if err != nil {
		log.Errorf("Unable to get material with id %d from course: %s", file_id, err.Error())
		handleApiError(c, err)
		return
	}

	c.File(file.URI)
	c.Status(http.StatusOK)
}

func (f *PublicController) DeleteMaterialFromCourse(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("course_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `course_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	file_id, err := strconv.Atoi(c.Param("file_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `file_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	err = coursematerial.DeleteMaterialFromCourse(f.Database, course_id, file_id)
	if err != nil {
		log.Errorf("Unable to delete file with id %d from course with id %d", file_id, course_id)
		handleApiError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (f *PublicController) DeleteUser(c *gin.Context) {
	role_id := c.MustGet("CookieRoleId").(int)

	if !AuthorizeAdmin(role_id) {
		handleApiError(c, errs.ErrNotAdmin)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	_, err = dbi.GetUserById(f.Database, id)
	if err != nil {
		log.Errorf("Unable to get user with id %d", id)
		handleApiError(c, err)
		return
	}

	err = dbi.DeleteUser(f.Database, id)
	if err != nil {
		log.Errorf("Unable to delete user from db: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetUserByCookie(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	if !AuthorizeUser(role_id) {
		handleApiError(c, errs.ErrNotUser)
		return
	}

	user, err := dbi.GetUserById(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get user with id %d", user_id)
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}

func (f *PublicController) GetUserById(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	if !AuthorizeUser(role_id) {
		handleApiError(c, errs.ErrNotUser)
		return
	}

	user_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	user, err := dbi.GetUserById(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get user with id %d", user_id)
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}

func (f *PublicController) GetAllAppointments(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)

	pCon := &calender.PublicController{Database: f.Database}
	appointments, err := pCon.GetAllAppointments(user_id)
	if err != nil {
		log.Errorf("Unable to get all appointments from user: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, appointments)
}

func (f *PublicController) AddCourseToCalender(c *gin.Context) {
	role_id := c.MustGet("CookieRoleId").(int)
	user_id := c.MustGet("CookieUserId").(int)

	var j map[string]interface{}

	if err := c.BindJSON(&j); err != nil {
		log.Errorf("Unable to bind json: %s\n", err.Error())
		// NOTE: `BindJSON` sets the return status arleady
		return
	}
	date, err := time.Parse(time.RFC3339, j["date"].(string))
	if err != nil {
		log.Error("unable to convert string to time.Time")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	duration, err := strconv.ParseInt(j["duration"].(string), 10, 32)
	if err != nil {
		log.Error("unable to convert string to int8")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	location, ok := j["location"].(string)
	if !ok {
		log.Error("unable to convert location to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	online, err := strconv.ParseInt(j["online"].(string), 10, 8)
	if err != nil {
		log.Error("unable to convert string to int8")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	courseId, err := strconv.ParseInt(j["courseId"].(string), 10, 64)
	if err != nil {
		log.Error("unable to convert string to int64")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	// authorization
	course_role, err := course.GetCourseRole(f.Database, user_id, int(courseId))
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	pCon := &calender.PublicController{Database: f.Database}
	_, err = pCon.AddCourseToCalender(date, int(duration), null.StringFrom(location), int8(online), int(courseId))
	if err != nil {
		log.Errorf("Unable to add course to calendar: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (f *PublicController) DeactivateCourseInCalender(c *gin.Context) {
	role_id := c.MustGet("CookieRoleId").(int)
	user_id := c.MustGet("CookieUserId").(int)

	var j map[string]interface{}

	if err := c.BindJSON(&j); err != nil {
		log.Errorf("Unable to bind json: %s", err.Error())
		// NOTE: `BindJSON` sets the return status arleady
		return
	}

	appointment_id, err := strconv.Atoi(j["appointment_id"].(string))
	if err != nil {
		log.Error("unable to convert string to int")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	// authorization
	appointment, err := models.FindAppointment(context.Background(), f.Database, appointment_id)
	if err != nil {
		log.Errorf("Unable to get appointment from id: %s", err.Error())
		handleApiError(c, err)
		return
	}
	course_role, err := course.GetCourseRole(f.Database, user_id, appointment.CourseID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	pCon := &calender.PublicController{Database: f.Database}
	err = pCon.DeactivateCourseInCalender(appointment_id)
	if err != nil {
		log.Errorf("Unable to deactivate course in calendar: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (f *PublicController) SearchCourse(c *gin.Context) {
	searchterm, ok := c.GetQuery("searchterm")
	if !ok {
		handleApiError(c, errs.ErrNoQuery)
		return
	}

	courses, err := course.SearchCourse(f.Database, searchterm)
	if err != nil {
		log.Errorf("Unable to search course: %s\n", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courses)
}

func (f *PublicController) CreateExam(c *gin.Context) {
	role_id := c.MustGet("CookieRoleId").(int)

	var newExam models.Exam

	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
		handleApiError(c, errs.ErrRawData)
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Errorf("Unable to unmarshal the json body: %+v", raw)
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	name, ok := j["name"].(string)
	if !ok {
		log.Error("unable to convert name to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	description, ok := j["description"].(string)
	if !ok {
		log.Error("unable to convert description to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	dateStr, ok := j["date"].(string)
	if !ok {
		log.Error("unable to convert date to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `date` to time.Time: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	durationStr, ok := j["duration"].(string)
	if !ok {
		log.Error("unable to convert duration to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `duration` to int: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	courseIdStr, ok := j["course_id"].(string)
	if !ok {
		log.Error("unable to convert course_id to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	courseId, err := strconv.Atoi(courseIdStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `course_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	creatorId := c.MustGet("CookieUserId").(int)

	course_role, err := course.GetCourseRole(f.Database, creatorId, courseId)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	onlineStr, ok := j["online"].(string)
	if !ok {
		log.Error("unable to convert online to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	online, err := strconv.Atoi(onlineStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `online` to int: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
	}

	location, ok := j["location"].(string)
	if !ok {
		log.Error("unable to convert location to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	registerDeadlineStr, ok := j["register_deadline"].(string)
	if !ok {
		log.Error("unable to convert register_deadline to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	registerDeadline, err := time.Parse(time.RFC3339, registerDeadlineStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `register_deadline` to time.Time: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	deregisterDeadlineStr, ok := j["deregister_deadline"].(string)
	if !ok {
		log.Error("unable to convert deregister_deadline to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	deregisterDeadline, err := time.Parse(time.RFC3339, deregisterDeadlineStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `deregister_deadline` to time.Time: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	id, err := pCtrl.CreateExam(name, description, date, duration, courseId, creatorId, int8(online), null.StringFrom(location), null.TimeFrom(registerDeadline), null.TimeFrom(deregisterDeadline))
	if err != nil {
		log.Errorf("Unable to create exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	newExam.ID = id

	c.IndentedJSON(http.StatusCreated, newExam)
}

func (f *PublicController) EditExam(c *gin.Context) {
	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
		handleApiError(c, errs.ErrRawData)
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Errorf("Unable to unmarshal the json body: %+v", raw)
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	name, ok := j["name"].(string)
	if !ok {
		log.Error("unable to convert name to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	description, ok := j["description"].(string)
	if !ok {
		log.Error("unable to convert description to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	dateStr, ok := j["date"].(string)
	if !ok {
		log.Error("unable to convert date to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	var date time.Time
	if dateStr != "" {
		date, err = time.ParseInLocation(time.RFC3339, dateStr, time.Local)
		if err != nil {
			log.Errorf("Unable to convert parameter `date` to time.Time: %s", err.Error())
			handleApiError(c, errs.ErrBodyConversion)
			return
		}
		date = date.Local()
	}
	durationStr, ok := j["duration"].(string)
	if !ok {
		log.Error("unable to convert duration to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	var duration int
	if durationStr != "" {
		duration, err = strconv.Atoi(durationStr)
		if err != nil {
			log.Errorf("Unable to convert parameter `duration` to int: %s", err.Error())
			handleApiError(c, errs.ErrBodyConversion)
			return
		}
	}

	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	userId := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	co, err := pCtrl.GetCourseFromExam(examId)
	if err != nil {
		log.Errorf("Unable to get course from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, userId, co.ID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}
	onlineStr, ok := j["online"].(string)
	if !ok {
		log.Error("unable to convert online to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	var online null.Int8
	if onlineStr != "" {
		onlineInt, err := strconv.Atoi(onlineStr)
		if err != nil {
			log.Errorf("Unable to convert parameter `online` to int: %s", err.Error())
			handleApiError(c, errs.ErrBodyConversion)
		}
		online.Int8 = int8(onlineInt)
		online.Valid = true
	} else {
		online.Valid = false
	}

	location, ok := j["location"].(string)
	if !ok {
		log.Error("unable to convert location to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	registerDeadlineStr, ok := j["register_deadline"].(string)
	if !ok {
		log.Error("unable to convert register_deadline to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	var registerDeadline time.Time
	if registerDeadlineStr != "" {
		registerDeadline, err = time.ParseInLocation(time.RFC3339, registerDeadlineStr, time.Local)
		if err != nil {
			log.Errorf("Unable to convert parameter `register_deadline` to time.Time: %s", err.Error())
			handleApiError(c, errs.ErrBodyConversion)
			return
		}
		registerDeadline = registerDeadline.Local()
	}

	deregisterDeadlineStr, ok := j["deregister_deadline"].(string)
	if !ok {
		log.Error("unable to convert register_deadline to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	var deregisterDeadline time.Time
	if deregisterDeadlineStr != "" {
		deregisterDeadline, err = time.ParseInLocation(time.RFC3339, deregisterDeadlineStr, time.Local)
		if err != nil {
			log.Errorf("Unable to convert parameter `deregister_deadline` to time.Time: %s", err.Error())
			handleApiError(c, errs.ErrBodyConversion)
			return
		}
		deregisterDeadline = deregisterDeadline.Local()
	}

	err = pCtrl.EditExam(name, description, date, duration, examId, online, null.StringFrom(location), null.TimeFrom(registerDeadline), null.TimeFrom(deregisterDeadline))
	if err != nil {
		log.Errorf("Unable to edit exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (f *PublicController) UploadExamFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	co, err := pCtrl.GetCourseFromExam(id)
	if err != nil {
		log.Errorf("Unable to get course from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, co.ID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseAdmin(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	if c.ContentType() == "text/plain" {
		var file _file
		if err := c.BindJSON(&file); err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			// NOTE: `BindJSON` sets the return status arleady
			return
		}

		err = pCtrl.UploadExamFile(file.Name, file.Uri, user_id, id, false, nil, 0)
		if err != nil {
			log.Errorf("Unable to create Exam-URI: %s", err.Error())
			handleApiError(c, err)
			return
		}
	} else {
		file, err := c.FormFile("file")
		if err != nil {
			log.Errorf("No file found in request: %s", err.Error())
			log.Error(err)
			handleApiError(c, errs.ErrNoFileInRequest)
			return
		}

		fi, err := file.Open()
		if err != nil {
			log.Errorf("Unable to open file: %s", err.Error())
			log.Error(err)
			handleApiError(c, err)
			return
		}
		err = pCtrl.UploadExamFile(file.Filename, "", user_id, id, true, fi, int(file.Size))
		if err != nil {
			log.Errorf("Unable to create ExamFile: %s", err.Error())
			handleApiError(c, err)
			return
		}
	}

	c.Status(http.StatusCreated)
}

func (f *PublicController) GetExamById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	ex, err := pCtrl.GetExamByID(id)
	if err != nil {
		log.Errorf("Unable to get exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	userId := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_role, err := course.GetCourseRole(f.Database, userId, ex.CourseID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}

	c.IndentedJSON(http.StatusOK, ex)
}

func (f *PublicController) GetRegisteredExamsFromUser(c *gin.Context) {
	userId := c.MustGet("CookieUserId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetRegisteredExamsFromUser(userId)
	if err != nil {
		log.Errorf("Unable to get exams from user: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetUnregisteredExamsFromUser(c *gin.Context) {
	userId := c.MustGet("CookieUserId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetUnregisteredExams(userId)
	if err != nil {
		log.Errorf("Unable to get exams: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetExamsFromCourse(c *gin.Context) {
	courseId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	course_role, err := course.GetCourseRole(f.Database, user_id, courseId)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}
	exams, err := pCtrl.GetExamsFromCourse(courseId)
	if err != nil {
		log.Errorf("Unable to get exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetAttendedExamsFromUser(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetAttendedExamsFromUser(user_id)
	if err != nil {
		log.Errorf("Unable to get exams from user: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetPassedExamsFromUser(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetPassedExamsFromUser(user_id)
	if err != nil {
		log.Errorf("Unable to get exams from user: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) GetCreatedFromUser(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	exams, err := pCtrl.GetCreatedExamsFromUser(user_id)
	if err != nil {
		log.Errorf("Unable to get exams from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, exams)
}

func (f *PublicController) RegisterToExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'exam_id' to an int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	userId := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	co, err := pCtrl.GetCourseFromExam(examId)
	if err != nil {
		log.Errorf("Unable to get course from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}
	course_role, err := course.GetCourseRole(f.Database, userId, co.ID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}

	user, err := pCtrl.RegisterToExam(userId, examId)
	if err != nil {
		log.Errorf("Unable to register user for exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}

func (f *PublicController) DeregisterFromExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'exam_id' to an int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	userId := c.MustGet("CookieUserId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	err = pCtrl.DeregisterFromExam(userId, examId)
	if err != nil {
		log.Errorf("Unable to deregister user from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetFileFromExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	userId := c.MustGet("CookieUserId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	if check, err := f.AuthorizeUserHasExam(userId, examId); !check {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusForbidden)
		return
	} else if err != nil {
		handleApiError(c, err)
		return
	}

	file, err := pCtrl.GetFileFromExam(examId)
	if err != nil {
		log.Errorf("Unable to get file from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	err = pCtrl.AttendExam(examId, userId)
	if err != nil {
		log.Errorf("Unable to get file from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.File(file[0].URI)
	c.Status(http.StatusOK)
}

func (f *PublicController) SubmitAnswerToExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	userId := c.MustGet("CookieUserId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	if check, err := f.AuthorizeUserHasExam(userId, examId); !check {
		log.Infof("User is not authorized: %s", err.Error())
		c.Status(http.StatusForbidden)
		return
	}

	if c.ContentType() == "text/plain" {
		var file _file
		if err := c.BindJSON(&file); err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			// NOTE: `BindJSON` sets the return status arleady
			return
		}

		err = pCtrl.SubmitAnswer(file.Name, file.Uri, examId, userId, false, nil, 0)
		if err != nil {
			log.Errorf("Unable to submit answer: %s", err.Error())
			handleApiError(c, err)
			return
		}
	} else {
		file, err := c.FormFile("file")
		if err != nil {
			log.Errorf("No file found in request: %s", err.Error())
			log.Error(err)
			handleApiError(c, errs.ErrNoFileInRequest)
			return
		}

		fi, err := file.Open()
		if err != nil {
			log.Errorf("Unable to open file: %s", err.Error())
			handleApiError(c, err)
			return
		}
		err = pCtrl.SubmitAnswer(file.Filename, "", examId, userId, true, fi, int(file.Size))
		if err != nil {
			log.Errorf("Unable to submit answer: %s", err.Error())
			handleApiError(c, err)
			return
		}
	}

	err = pCtrl.AttendExam(examId, userId)
	if err != nil {
		log.Errorf("Unable to attend user to exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (f *PublicController) GetRegisteredUsersFromExam(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'id' to an int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	creatorId := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	co, err := pCtrl.GetCourseFromExam(examId)
	if err != nil {
		log.Errorf("Unable to get course from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, creatorId, co.ID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	attendees, err := pCtrl.GetRegisteredUsersFromExam(examId, creatorId)
	if err != nil {
		log.Errorf("Unable to fetch attendees from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, attendees)
}

func (f *PublicController) GetAttendeesFromExam(c *gin.Context) {
	creatorId := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	pCtrl := exam.PublicController{Database: f.Database}

	examId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'id' to an int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	co, err := pCtrl.GetCourseFromExam(examId)
	if err != nil {
		log.Errorf("Unable to get course from exam: %s", err)
		handleApiError(c, err)
		return
	}

	course_role_id, err := course.GetCourseRole(f.Database, creatorId, co.ID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err)
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseModerator(course_role_id, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	attendees, err := pCtrl.GetAttendeesFromExam(examId, creatorId)
	if err != nil {
		log.Errorf("Unable to fetch attendees from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, attendees)
}

func (f *PublicController) GetFileFromAttendee(c *gin.Context) {
	attendeeId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	examId, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter 'exam_id' to an int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	userId := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	co, err := pCtrl.GetCourseFromExam(examId)
	if err != nil {
		log.Errorf("Unable to get course from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, userId, co.ID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	file, err := pCtrl.GetAnswerFromAttendee(attendeeId, examId)
	if err != nil {
		log.Errorf("Unable to get file from answer: %s", err.Error())
		handleApiError(c, err)
		return
	}
	c.File(file.URI)
	c.Status(http.StatusOK)
}

func (f *PublicController) GradeAnswer(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `exam_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	creatorId := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `user_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	co, err := pCtrl.GetCourseFromExam(examId)
	if err != nil {
		log.Errorf("Unable to get course from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, creatorId, co.ID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}
	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
		handleApiError(c, errs.ErrRawData)
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Errorf("Unable to unmarshal the json body: %+v", raw)
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	gradeStr, ok := j["grade"].(string)
	if !ok {
		log.Error("unable to convert grade to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	grade, err := strconv.Atoi(gradeStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `grade` to int: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	passedStr, ok := j["passed"].(string)
	if !ok {
		log.Error("unable to convert passed to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	passed, err := strconv.Atoi(passedStr)
	if err != nil {
		log.Errorf("Unable to convert parameter `passed` to int: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	feedback, ok := j["feedback"].(string)
	if !ok {
		log.Error("unable to convert feedback to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	err = pCtrl.GradeAnswer(examId, creatorId, userId, null.IntFrom(grade), null.Int8From(int8(passed)), null.StringFrom(feedback))
	if err != nil {
		log.Errorf("Unable to grade answer: %s", err.Error())
		handleApiError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func (f *PublicController) SetAttended(c *gin.Context) {
	examId, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `exam_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}
	attendeeId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `user_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	pCtrl := exam.PublicController{Database: f.Database}
	co, err := pCtrl.GetCourseFromExam(examId)
	if err != nil {
		log.Errorf("Unable to get course from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	userId := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_role, err := course.GetCourseRole(f.Database, userId, co.ID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	err = pCtrl.SetAttended(examId, attendeeId)
	if err != nil {
		log.Errorf("Unable to set user's exam to attended: %s", err.Error())
		handleApiError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func (f *PublicController) DeleteExam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	userId := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	pCtrl := exam.PublicController{Database: f.Database}
	co, err := pCtrl.GetCourseFromExam(id)
	if err != nil {
		log.Errorf("Unable to get course from exam: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, userId, co.ID)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseAdmin(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseAdmin)
		return
	}
	ex, err := pCtrl.DeleteExam(id)
	// Return Status and Data in JSON-Format
	if err != nil {
		log.Errorf("Unable to delete course: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, ex)
}
func (f *PublicController) GetSubmission(c *gin.Context) {
	submission_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}
	submission, err := course.GetSubmission(f.Database, submission_id)
	if err != nil {
		log.Errorf("Unable to get submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, submission)
}

func (f *PublicController) CreateSubmission(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
		handleApiError(c, errs.ErrRawData)
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Errorf("Unable to unmarshal the json body: %+v", raw)
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	name, ok := j["name"].(string)
	if !ok {
		log.Error("unable to convert name to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	deadline, ok := j["deadline"].(string)
	if !ok {
		log.Error("unable to convert deadline to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	visible_from, ok := j["visible_from"].(string)
	if !ok {
		log.Error("unable to convert visible_from to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	max_filesize, err := strconv.Atoi(j["max_filesize"].(string))
	if err != nil {
		log.Error("unable to convert maxfilesize to int")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	id, err := course.CreateSubmission(f.Database, name, deadline, course_id, max_filesize, visible_from)
	if err != nil {
		log.Errorf("Unable to create submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, id)
}

func (f *PublicController) DeleteSubmission(c *gin.Context) {
	submission_id, err := strconv.Atoi(c.Param("submission_id"))
	if err != nil {
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	submission_id, err = course.DeleteSubmission(f.Database, submission_id)
	if err != nil {
		log.Errorf("Unable to delete submission: %s\n", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, submission_id)
}

func (f *PublicController) EditSubmissionById(c *gin.Context) {
	submission_id, err := strconv.Atoi(c.Param("submission_id"))
	if err != nil {
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := course.GetCourseIdBySubmission(f.Database, submission_id)
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}
	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
		handleApiError(c, err)
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Errorf("Unable to unmarshal the json body: %+v", raw)
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	name, ok := j["name"].(string)
	if !ok {
		log.Error("unable to convert name to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	deadline, ok := j["deadline"].(string)
	if !ok {
		log.Error("unable to convert deadline to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	visible_from, ok := j["visible_from"].(string)
	if !ok {
		log.Error("unable to convert visible_from to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	max_filesize, err := strconv.Atoi(j["max_filesize"].(string))
	if err != nil {
		log.Error("unable to convert maxfilesize to int")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	_, err = course.EditSubmission(f.Database, submission_id, name, deadline, max_filesize, visible_from)
	if err != nil {
		log.Errorf("Unable to update submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, submission_id)
}

func (f *PublicController) GetSubmissionFromUser(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)

	users, err := course.GetSubmissionsFromUser(f.Database, user_id)
	if err != nil {
		log.Errorf("Unable to get submissions from user: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, users)
}

func (f *PublicController) CreateSubmissionHasFiles(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	submission_id, err := strconv.Atoi(c.Param("submission_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_id, err := course.GetCourseIdBySubmission(f.Database, submission_id)
	if err != nil {
		log.Errorf("Unable to get `course_id` by submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	if c.ContentType() == "text/plain" {
		var file _file
		if err := c.BindJSON(&file); err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			// NOTE: `BindJSON` sets the return status arleady
			return
		}

		err = course.CreateSubmissionHasFiles(f.Database, submission_id, file.Name, file.Uri, user_id, false, nil, 0)
		if err != nil {
			log.Errorf("Unable to add file to submission: %s", err.Error())
			handleApiError(c, err)
			return
		}
	} else {
		file, err := c.FormFile("file")
		if err != nil {
			log.Errorf("No file found in request: %s", err.Error())
			log.Error(err)
			handleApiError(c, errs.ErrNoFileInRequest)
			return
		}

		fi, err := file.Open()
		if err != nil {
			log.Errorf("Unable to open file: %s", err.Error())
			handleApiError(c, err)
			return
		}

		err = course.CreateSubmissionHasFiles(f.Database, submission_id, file.Filename, "", user_id, true, fi, int(file.Size))
		if err != nil {
			log.Errorf("Unable to add file to submission: %s", err.Error())
			handleApiError(c, err)
			return
		}
	}

	c.Status(http.StatusCreated)
}

func (f *PublicController) DeleteSubmissionHasFiles(c *gin.Context) {

	submission_id, err := strconv.Atoi(c.Param("submission_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}
	file_id, err := strconv.Atoi(c.Param("file_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := course.GetCourseIdBySubmission(f.Database, submission_id)
	if err != nil {
		log.Errorf("Unable to get `course_id` by submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}

	err = course.DeleteSubmissionHasFiles(f.Database, submission_id, file_id)
	if err != nil {
		log.Errorf("Unable to delete file from submission: %s", err.Error())
		handleApiError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetUserSubmission(c *gin.Context) {
	user_submission_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}
	user_submission, err := course.GetUserSubmission(f.Database, user_submission_id)
	if err != nil {
		log.Errorf("Unable to get user submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, user_submission)
}

func (f *PublicController) CreateUserSubmission(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	submission_id, err := strconv.Atoi(c.Param("submission_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}
	course_id, err := course.GetCourseIdBySubmission(f.Database, submission_id)
	if err != nil {
		log.Errorf("Unable to `course_id` by submission: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}
	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}
	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}
	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
		handleApiError(c, errs.ErrRawData)
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Errorf("Unable to unmarshal the json body: %+v", raw)
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	name, ok := j["name"].(string)
	if !ok {
		log.Error("unable to convert name to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	ignores_deadline, ok := j["ignores_deadline"].(string)
	if !ok {
		log.Error("unable to convert ignores_deadline to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	ignores_deadlineint, err := strconv.Atoi(ignores_deadline)
	if err != nil {
		log.Error("unable to convert ignores_deadline to int")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	user_submission_id, err := course.CreateUserSubmission(f.Database, name, 9999, submission_id, int8(ignores_deadlineint))
	if err != nil {
		log.Errorf("Unable to add file to submission: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	c.IndentedJSON(http.StatusCreated, user_submission_id)
}

func (f *PublicController) DeleteUserSubmission(c *gin.Context) {
	user_submission_id, err := strconv.Atoi(c.Param("usersubmission_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := course.GetCourseIdByUserSubmission(f.Database, user_submission_id)
	if err != nil {
		log.Errorf("Unable to get `course_id` by submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}

	user_submission_id, err = course.DeleteUserSubmission(f.Database, user_submission_id, user_id)
	if err != nil {
		log.Errorf("Unable to delete user submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, user_submission_id)
}

func (f *PublicController) CreateUserSubmissionHasFiles(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	user_submission_id, err := strconv.Atoi(c.Param("usersubmission_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `usersubmission_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_id, err := course.GetCourseIdByUserSubmission(f.Database, user_submission_id)
	if err != nil {
		log.Errorf("Unable to get `course_id` by submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}

	if c.ContentType() == "text/plain" {

		var file _file
		if err := c.BindJSON(&file); err != nil {
			log.Errorf("Unable to bind json: %s", err.Error())
			// NOTE: `BindJSON` sets the return status arleady
			return
		}

		err = course.CreateUserSubmissionHasFiles(f.Database, user_submission_id, file.Name, file.Uri, user_id, false, nil, 0)
		if err != nil {
			log.Errorf("Unable to add file to user submission: %s", err.Error())
			handleApiError(c, err)
			return
		}
	} else {
		file, err := c.FormFile("file")
		if err != nil {
			log.Errorf("No file found in request: %s", err.Error())
			log.Error(err)
			handleApiError(c, errs.ErrNoFileInRequest)
			return
		}

		fi, err := file.Open()
		if err != nil {
			log.Errorf("Unable to open file: %s", err.Error())
			handleApiError(c, err)
			return
		}

		err = course.CreateUserSubmissionHasFiles(f.Database, user_submission_id, file.Filename, "", user_id, true, fi, int(file.Size))
		if err != nil {
			log.Errorf("Unable to add file to user submission: %s", err.Error())
			handleApiError(c, err)
			return
		}
	}

	c.Status(http.StatusCreated)
}

func (f *PublicController) DeleteUserSubmissionHasFiles(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	user_submission_id, err := strconv.Atoi(c.Param("usersubmission_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `usersubmission_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_id, err := course.GetCourseIdByUserSubmission(f.Database, user_submission_id)
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, err)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}

	file_id, err := strconv.Atoi(c.Param("file_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `file_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	if !AuthorizeModerator(role_id) {
		handleApiError(c, errs.ErrNotModerator)
		return
	}

	err = course.DeleteUserSubmissionHasFiles(f.Database, user_submission_id, file_id, user_id)
	if err != nil {
		log.Errorf("Unable to delete file from user submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetSubmissionsFromCourse(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}
	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseUser(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseUser)
		return
	}

	submissions, err := course.GetSubmissionsFromCourse(f.Database, course_id)
	if err != nil {
		log.Errorf("Unable to get submissions from course: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, submissions)
}

func (f *PublicController) GradeUserSubmission(c *gin.Context) {
	user_submission_id, err := strconv.Atoi(c.Param("usersubmission_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `usersubmission_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	user_id := c.MustGet("CookieUserId").(int)
	role_id := c.MustGet("CookieRoleId").(int)

	course_id, err := course.GetCourseIdByUserSubmission(f.Database, user_submission_id)
	if err != nil {
		log.Errorf("Unable to get `course_id` by submission: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err.Error())
		handleApiError(c, err)
		return
	}

	if !AuthorizeCourseModerator(course_role, role_id) {
		handleApiError(c, errs.ErrNotCourseModerator)
		return
	}
	raw, err := c.GetRawData()
	if err != nil {
		log.Errorf("Unable to get raw data from request: %s", err.Error())
		handleApiError(c, errs.ErrRawData)
		return
	}

	var j map[string]interface{}
	err = json.Unmarshal(raw, &j)
	if err != nil {
		log.Errorf("Unable to unmarshal the json body: %+v", raw)
		handleApiError(c, errs.ErrBodyConversion)
		return
	}

	grade, ok := j["grade"].(string)
	if !ok {
		log.Error("unable to convert grade to string")
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	gradeint, err := strconv.Atoi(grade)
	if err != nil {
		log.Errorf("Unable to convert grade string to int: %s", err.Error())
		handleApiError(c, errs.ErrBodyConversion)
		return
	}
	// Deactivate Data from Database with Backend function
	err = course.GradeUserSubmission(f.Database, user_submission_id, gradeint)
	// Return Status and Data in JSON-Format
	if err != nil {
		handleApiError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (f *PublicController) GetUserSubmissionsFromSubmission(c *gin.Context) {
	submission_id, err := strconv.Atoi(c.Param("submission_id"))
	if err != nil {
		log.Errorf("Unable to convert parameter `submission_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	submissions, err := course.GetUserSubmissionsFromSubmission(f.Database, submission_id)
	if err != nil {
		log.Errorf("Unable to get usersubmissions from submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, submissions)
}

func (f *PublicController) GetFileFromSubmission(c *gin.Context) {
	submission_id, err := strconv.Atoi(c.Param("submission_id"))
	if err != nil {
		log.Errorf("unable to convert parameter `submission_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	users, err := course.GetFileFromSubmission(f.Database, submission_id)
	if err != nil {
		log.Errorf("unable to get file from submission: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, users)
}

func (f *PublicController) GetFileFromUserSubmission(c *gin.Context) {
	user_submission_id, err := strconv.Atoi(c.Param("usersubmission_id"))
	if err != nil {
		log.Errorf("unable to convert parameter `usersubmission_id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	users, err := course.GetFileFromUserSubmission(f.Database, user_submission_id)
	if err != nil {
		log.Errorf("Unable to get users in course: %s", err.Error())
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, users)
}

func (f *PublicController) GetUserCourseRole(c *gin.Context) {
	user_id := c.MustGet("CookieUserId").(int)

	course_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("unable to convert parameter `id` to int: %s", err.Error())
		handleApiError(c, errs.ErrParameterConversion)
		return
	}

	course_role_id, err := course.GetCourseRole(f.Database, user_id, course_id)
	if err != nil {
		log.Errorf("Unable to get course role: %s", err)
		handleApiError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, course_role_id)
}
