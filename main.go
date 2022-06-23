package main

import (
	"database/sql"

	"learningbay24.de/backend/api"
	"learningbay24.de/backend/config"
	"learningbay24.de/backend/dbi"

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

func main() {
	config.InitConfig()
	config.InitLogger()
	db := config.SetupDbHandle()
	applyMigrations(db)
	setupEnvironment(db)

	pCtrl := api.PublicController{Database: db}
	router := gin.Default()
	router.Use(CORSMiddleware())

	router.GET("/courses/:id", pCtrl.GetCourseById)
	router.GET("/users/courses", pCtrl.GetCoursesFromUser)
	router.GET("/courses/:id/users", pCtrl.GetUsersInCourse)
	router.DELETE("/courses/:id", pCtrl.DeleteCourse)
	router.DELETE("/courses/:id/:user_id", pCtrl.DeleteUserFromCourse)
	router.POST("/login", pCtrl.Login)
	router.POST("/logout", pCtrl.Logout)
	router.POST("/register", pCtrl.Register)
	router.POST("/courses", pCtrl.CreateCourse)
	router.POST("/courses/:id", pCtrl.EnrollUser)
	router.POST("/courses/:id/files", pCtrl.UploadMaterial)
	router.GET("/submissions/:id", pCtrl.GetSubmission)
	router.GET("/courses/:id/submissions", pCtrl.GetSubmissionsFromCourse)
	router.POST("/courses/:id/submissions", pCtrl.CreateSubmission)
	router.GET("/courses/:id/submissions/:submission_id/usersubmissions", pCtrl.GetUserSubmissionsFromSubmission)
	router.PATCH("/courses/:id/submissions/usersubmissions/:usersubmission_id/grade", pCtrl.GradeUserSubmission)
	router.DELETE("courses/:id/submissions/:submission_id", pCtrl.DeleteSubmission)
	router.PATCH("courses/:id/submissions/:submission_id", pCtrl.EditSubmissionById)
	router.GET("/courses/:id/submissions/:submission_id/files", pCtrl.GetFileFromSubmission)
	router.POST("/courses/:id/submissions/:submission_id/files", pCtrl.CreateSubmissionHasFiles)
	router.DELETE("/courses/:id/submissions/:submission_id/files/:file_id", pCtrl.DeleteSubmissionHasFiles)
	router.POST("/courses/:id/submissions/:submission_id/usersubmissions", pCtrl.CreateUserSubmission)
	router.DELETE("/courses/:id/submissions/usersubmissions/:usersubmission_id", pCtrl.DeleteUserSubmission)
	router.GET("/courses/:id/submissions/usersubmissions/:usersubmission_id/files", pCtrl.GetFileFromUserSubmission)
	router.POST("/courses/:id/submissions/usersubmissions/:usersubmission_id/files", pCtrl.CreateUserSubmissionHasFiles)
	router.DELETE("/courses/:id/submissions/usersubmissions/:usersubmission_id/files/:file_id", pCtrl.DeleteUserSubmissionHasFiles)
	router.GET("/courses/:id/files", pCtrl.GetMaterialsFromCourse)
	router.GET("/courses/:id/files/:file_id", pCtrl.GetMaterialFromCourse)
	router.GET("/courses/search", pCtrl.SearchCourse)
	router.PATCH("/courses/:id", pCtrl.EditCourseById)
	router.DELETE("/users/:id", pCtrl.DeleteUser)
	router.GET("/users/cookie", pCtrl.GetUserByCookie)
	router.GET("/users/:id", pCtrl.GetUserById)
	router.POST("/exams", pCtrl.CreateExam)
	router.PATCH("/exams/:id/edit", pCtrl.EditExam)
	router.POST("/exams/:id/files", pCtrl.UploadExamFile)
	router.GET("/exams/:id", pCtrl.GetExamById)
	router.GET("/courses/:id/exams", pCtrl.GetExamsFromCourse)
	router.GET("/users/exams/registered", pCtrl.GetRegisteredExamsFromUser)
	router.GET("/users/exams/unregistered", pCtrl.GetUnregisteredExamsFromUser)
	router.GET("/users/exams/attended", pCtrl.GetAttendedExamsFromUser)
	router.GET("/users/exams/passed", pCtrl.GetPassedExamsFromUser)
	router.GET("/users/exams/created", pCtrl.GetCreatedFromUser)
	router.POST("/users/exams/:id", pCtrl.RegisterToExam)
	router.DELETE("/users/exams/:id", pCtrl.DeregisterFromExam)
	router.PATCH("/users/:user_id/exams/:exam_id/attend", pCtrl.SetAttended)
	router.GET("/exams/:id/files", pCtrl.GetFileFromExam)
	router.POST("/users/exams/:id/submit", pCtrl.SubmitAnswerToExam)
	router.GET("/exams/:id/users", pCtrl.GetRegisteredUsersFromExam)
	router.GET("exams/:id/users/attended", pCtrl.GetAttendeesFromExam)
	// NOTE: `usersx` is used as `users` seems to cause problems for this particular route
	router.GET("/usersx/:id/exams/:exam_id/files", pCtrl.GetFileFromAttendee)
	router.PATCH("/users/:user_id/exams/:exam_id/grade", pCtrl.GradeAnswer)
	router.DELETE("/exams/:id", pCtrl.DeleteExam)
	router.GET("/courses/appointments", pCtrl.GetAllAppointments)
	router.POST("/appointments/add", pCtrl.AddCourseToCalender)
	router.DELETE("/appointments", pCtrl.DeactivateCourseInCalender)
	router.GET("/users/submissions", pCtrl.GetSubmissionFromUser)
	router.GET("/users/submissions/:id", pCtrl.GetUserSubmission)

	router.Run("0.0.0.0:8080")
}
