package main

import (
	"github.com/gin-gonic/gin"
	"learningbay24.de/backend/api"
	"learningbay24.de/backend/config"
)

func main() {
	config.InitConfig()
	// TODO: do something with the db handle
	db := config.SetupDbHandle()
	pCtrl := api.PublicController{Database: db}
	router := gin.Default()
	router.GET("/courses/:id", pCtrl.GetCourseById)
	router.GET("/courses/:user_id", pCtrl.GetUserCourses)
	router.GET("/courses/:id/users", pCtrl.GetUsersInCourse)
	router.DELETE("/courses/:id", pCtrl.DeleteCourse)
	router.DELETE("/courses/:id/:user_id", pCtrl.DeleteUserFromCourse)
	router.POST("/courses", pCtrl.CreateCourse)
	router.POST("/courses/:id/:user_id", pCtrl.EnrollUser)
	router.PATCH("/courses/:id", pCtrl.UpdateCourseById)
	router.Run("0.0.0.0:8080")
}
