package main

import (
	"database/sql"
	"log"

	"learningbay24.de/backend/api"
	"learningbay24.de/backend/config"

	"github.com/gin-gonic/gin"
	"github.com/rubenv/sql-migrate"
)

func applyMigrations(db *sql.DB) {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	n, err := migrate.Exec(db, "mysql", migrations, migrate.Up)
	if err != nil {
		log.Fatal("Unable to apply migrations. Aborting.")
	}
	log.Printf("Applied %d migrations\n", n)
}

func main() {
	config.InitConfig()
	db := config.SetupDbHandle()
	applyMigrations(db)

	pCtrl := api.PublicController{Database: db}
	router := gin.Default()

	router.GET("/courses/:id", pCtrl.GetCourseById)
	router.GET("/users/:user_id/courses", pCtrl.GetCoursesFromUser)
	router.GET("/courses/:id/users", pCtrl.GetUsersInCourse)
	router.DELETE("/courses/:id", pCtrl.DeleteCourse)
	router.DELETE("/courses/:id/:user_id", pCtrl.DeleteUserFromCourse)
	router.POST("/login", pCtrl.Login)
	router.POST("/register", pCtrl.Register)
	router.POST("/courses", pCtrl.CreateCourse)
	router.POST("/courses/:id/:user_id", pCtrl.EnrollUser)
	router.PATCH("/courses/:id", pCtrl.UpdateCourseById)

	router.Run("0.0.0.0:8080")
}
