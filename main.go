package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"learningbay24.de/backend/api"
	"learningbay24.de/backend/config"
	"learningbay24.de/backend/dbi"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
)

func applyMigrations(db *sql.DB) {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	n, err := migrate.Exec(db, "mysql", migrations, migrate.Up)
	if err != nil {
		log.Fatalf("Unable to apply migrations: %s. Aborting.", err.Error())
	}

	log.Infof("Applied %d migrations\n", n)
}

func setupEnvironment(db *sql.DB) {
	// always try to set up admin user
	if err := dbi.AddDefaultData(db); err != nil {
		log.Info("Unable to insert default data. This could be due to default data already being inserted!")
		log.Info(err)
	}

	if config.Conf.Environment != "development" {
		return
	}

	dbi.AddDummyData(db)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://learningbay24.de")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
		} else {
			c.Next()
		}
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		flog := log.WithFields(log.Fields{
			"context": "auth_middleware",
		})

		cookie := c.Request.Header.Get("Cookie")
		if cookie == "" {
			flog.Errorf("Unable to get cookie")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenString := strings.Split(cookie, "=")[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(config.Conf.Secrets.JWTSecret), nil
		})
		if err != nil {
			flog.Errorf("Error parsing token: %s", err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		data, ok := token.Claims.(jwt.MapClaims)["data"]
		if !ok {
			flog.Error("Unable to map id from data interface")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		datamap, ok := data.(map[string]interface{})
		if !ok {
			flog.Error("Unable to map id from data interface")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		id, err := strconv.Atoi(datamap["id"].(string))
		if err != nil {
			flog.Errorf("Unable to convert id to int: %s", err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		role_id, err := strconv.Atoi(datamap["role_id"].(string))
		if err != nil {
			flog.Errorf("Unable to convert role_id to int: %s", err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// ensure basic permissions
		if !api.AuthorizeUser(role_id) {
			flog.Error("Cookie's role_id does not have user permissions")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("CookieUserId", id)
		c.Set("CookieRoleId", role_id)
		c.Next()
	}
}

func main() {
	config.InitConfig()
	config.InitLogger()
	db := config.SetupDbHandle()
	applyMigrations(db)
	setupEnvironment(db)

	pCtrl := api.PublicController{Database: db}
	router := gin.Default()
	router.Use(CORSMiddleware())

	auth := router.Group("").Use(AuthMiddleware())
	{
		auth.GET("/courses/:id", pCtrl.GetCourseById)
		auth.DELETE("/courses/:id/:user_id", pCtrl.DeleteUserFromCourse)
		auth.GET("/courses/:id/users", pCtrl.GetUsersInCourse)
		auth.GET("/users/courses", pCtrl.GetCoursesFromUser)
		auth.DELETE("/courses/:id", pCtrl.DeleteCourse)
		auth.POST("/courses", pCtrl.CreateCourse)
		auth.POST("/courses/:id", pCtrl.EnrollUser)
		auth.PATCH("/courses/:id", pCtrl.EditCourseById)
		auth.POST("/logout", pCtrl.Logout)
		auth.POST("/register", pCtrl.Register)
		auth.POST("/courses/:id/files", pCtrl.UploadMaterial)
		auth.GET("/courses/:id/files", pCtrl.GetMaterialsFromCourse)
		auth.GET("/courses/:id/files/:file_id", pCtrl.GetMaterialFromCourse)
		auth.DELETE("/courses/:id/files/:file_id", pCtrl.DeleteMaterialFromCourse)
		auth.DELETE("/users/:id", pCtrl.DeleteUser)
		auth.GET("/users/cookie", pCtrl.GetUserByCookie)
		auth.GET("/users/:id", pCtrl.GetUserById)
		auth.GET("/courses/appointments", pCtrl.GetAllAppointments)
		auth.POST("/exams", pCtrl.CreateExam)
		auth.PATCH("/exams/:id/edit", pCtrl.EditExam)
		auth.POST("/exams/:id/files", pCtrl.UploadExamFile)
		auth.GET("/users/exams/registered", pCtrl.GetRegisteredExamsFromUser)
		auth.GET("/users/exams/unregistered", pCtrl.GetUnregisteredExamsFromUser)
		auth.GET("/courses/:id/exams", pCtrl.GetExamsFromCourse)
		auth.GET("/users/exams/attended", pCtrl.GetAttendedExamsFromUser)
		auth.GET("/users/exams/passed", pCtrl.GetPassedExamsFromUser)
		auth.GET("/users/exams/created", pCtrl.GetCreatedFromUser)
		auth.POST("/users/exams/:id", pCtrl.RegisterToExam)
		auth.DELETE("/users/exams/:id", pCtrl.DeregisterFromExam)
		auth.GET("/exams/:id/files", pCtrl.GetFileFromExam)
		auth.POST("/users/exams/:id/submit", pCtrl.SubmitAnswerToExam)
		auth.GET("/exams/:id/users", pCtrl.GetRegisteredUsersFromExam)
		auth.GET("/exams/:id/users/attended", pCtrl.GetAttendeesFromExam)
		auth.PATCH("/users/:user_id/exams/:exam_id/grade", pCtrl.GradeAnswer)
		auth.DELETE("/exams/:id", pCtrl.DeleteExam)
		auth.GET("/exams/:id", pCtrl.GetExamById)
		auth.PATCH("/users/:user_id/exams/:exam_id/attend", pCtrl.SetAttended)
		auth.GET("/usersx/:id/exams/:exam_id/files", pCtrl.GetFileFromAttendee)
		auth.POST("/courses/:id/submissions", pCtrl.CreateSubmission)
		auth.DELETE("/courses/:id/submissions/:submission_id", pCtrl.DeleteSubmission)
		auth.PATCH("/courses/submissions/:submission_id", pCtrl.EditSubmissionById)
		auth.GET("/users/submissions", pCtrl.GetSubmissionFromUser)
		auth.POST("/courses/submissions/:submission_id/files", pCtrl.CreateSubmissionHasFiles)
		auth.DELETE("/courses/submissions/:submission_id/files/:file_id", pCtrl.DeleteSubmissionHasFiles)
		auth.POST("/courses/submissions/:submission_id/usersubmissions", pCtrl.CreateUserSubmission)
		auth.DELETE("/courses/submissions/usersubmissions/:usersubmission_id", pCtrl.DeleteUserSubmission)
		auth.POST("/courses/:id/submissions/usersubmissions/:usersubmission_id/files", pCtrl.CreateUserSubmissionHasFiles)
		auth.DELETE("/courses/submissions/usersubmissions/:usersubmission_id/files/:file_id", pCtrl.DeleteUserSubmissionHasFiles)
		auth.GET("/courses/:id/submissions", pCtrl.GetSubmissionsFromCourse)
		auth.PATCH("/courses/submissions/usersubmissions/:usersubmission_id/grade", pCtrl.GradeUserSubmission)
	}

	router.POST("/login", pCtrl.Login)
	// TODO: add authorization => user has access to submission
	router.GET("/submissions/:id", pCtrl.GetSubmission)
	// TODO: add authorization => user
	router.GET("/courses/submissions/:submission_id/usersubmissions", pCtrl.GetUserSubmissionsFromSubmission)
	// TODO: add authorization => user
	router.GET("/courses/submissions/:submission_id/files", pCtrl.GetFileFromSubmission)
	// TODO: add authorization => user
	router.GET("/courses/submissions/usersubmissions/:usersubmission_id/files", pCtrl.GetFileFromUserSubmission)
	router.GET("/courses/search", pCtrl.SearchCourse)
	// TODO: add authorization?
	router.POST("/appointments/add", pCtrl.AddCourseToCalender)
	// TODO: add authorization?
	router.DELETE("/appointments", pCtrl.DeactivateCourseInCalender)
	// TODO: add authorization => user
	router.GET("/users/submissions/:id", pCtrl.GetUserSubmission)

	router.Run("0.0.0.0:8080")
}
