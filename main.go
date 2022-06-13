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
	if config.Conf.Environment != "development" {
		return
	}

	dbi.AddDummyData(db)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://learningbay24.de")
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
	router.POST("/register", pCtrl.Register)
	router.POST("/courses", pCtrl.CreateCourse)
	router.POST("/courses/:id/:user_id", pCtrl.EnrollUser)
	router.POST("/courses/:id/files", pCtrl.UploadMaterial)
	router.GET("/courses/:id/files", pCtrl.GetMaterialsFromCourse)
	router.GET("/courses/:id/files/:file_id", pCtrl.GetMaterialFromCourse)
	router.GET("/courses/search", pCtrl.SearchCourse)
	router.PATCH("/courses/:id", pCtrl.UpdateCourseById)
	router.DELETE("/users/:id", pCtrl.DeleteUser)
	// TODO: panics
	router.GET("/courses/appointments", pCtrl.GetAllAppointments)
	router.GET("/appointments", pCtrl.GetAppointments)
	router.GET("/users/courses/submissions", pCtrl.GetAllSubmissions)
	router.POST("/appointments/add", pCtrl.AddCourseToCalender)
	router.POST("/appointments/addsubmission", pCtrl.AddSubmissionToCalender)
	router.DELETE("/appointments/:id", pCtrl.DeactivateAppointment)
	router.DELETE("/appointments/:id/appointment/", pCtrl.DeactivateCourseInCalender)
	router.DELETE("/appointments/:id/exam", pCtrl.DeactivateExamInCalender)
	//router.GET("/users/cookie", pCtrl.GetUserByCookie)
	router.GET("/users/:id", pCtrl.GetUserById)

	router.Run("0.0.0.0:8080")
}
