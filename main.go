package main

import (
	"github.com/gin-gonic/gin"
	"learningbay24.de/backend/api"
	"learningbay24.de/backend/config"
)

func main() {
	config.InitConfig()
	db := config.SetupDbHandle()

	pCtrl := api.PublicController{Database: db}
	router := gin.Default()

	router.GET("/courses/:id", pCtrl.GetCourseById)
	router.GET("/users/:user_id/courses", pCtrl.GetCoursesFromUser)
	router.GET("/courses/:id/users", pCtrl.GetUsersInCourse)
	router.DELETE("/courses/:id", pCtrl.DeleteCourse)
	router.DELETE("/courses/:id/:user_id", pCtrl.DeleteUserFromCourse)
	router.POST("/courses", pCtrl.CreateCourse)
	router.POST("/courses/:id/:user_id", pCtrl.EnrollUser)
	router.PATCH("/courses/:id", pCtrl.UpdateCourseById)
	router.POST("/exams", pCtrl.CreateExam)

	router.Run("0.0.0.0:8080")
}
