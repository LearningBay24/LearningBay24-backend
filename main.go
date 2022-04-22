package main

import (
	"learningbay24.de/backend/config"
	"github.com/gin-gonic/gin"
	"learningbay24.de/backend/api"
)

func main() {
	config.InitConfig()
	// TODO: do something with the db handle
	_ = config.SetupDbHandle()
	router := gin.Default()
	router.GET("/courses/:id",api.GetCourseById)
	router.POST("/courses",api.CreateCourse)
	router.Run("0.0.0.0:8080")
}
