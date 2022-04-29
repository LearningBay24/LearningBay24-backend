package main

import (
	"github.com/gin-gonic/gin"
	"learningbay24.de/backend/api"
	"learningbay24.de/backend/config"
)

func main() {
	config.InitConfig()
	// TODO: do something with the db handle
	_ = config.SetupDbHandle()
	router := gin.Default()
	router.GET("/courses/:id",api.GetCourseById)
	router.GET("user/:user_id/courses",api.GetUserCourses)
	router.GET("courses/:id/users",api.GetUsersInCourse)
	router.DELETE("/courses/:id/delete",api.DeleteCourseById)
	router.DELETE("/courses/:id/delete/user/:user_id",api.DeleteUserFromCourse)
	router.POST("/courses/create",api.CreateCourse)
	router.POST("/courses/:id/enroll/user/:user_id",api.EnrollUser)
	router.PATCH("/courses/:id/update",api.UpdateCourseById)
	router.PATCH("/courses/:id/deactivate",api.DeactivateCourse)
	//router.GET("/courses", api.GetCourses)
	router.Run("0.0.0.0:8080")
}