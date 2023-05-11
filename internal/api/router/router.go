package router

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"scheduler/internal/api/handler"
)

func Get() *gin.Engine {
	r := gin.Default()

	// 路由定义
	taskAPI := r.Group("/api/tasks")
	{
		taskAPI.GET("", handler.GetTaskListHandler)
		taskAPI.POST("", handler.SubmitTaskHandler)
		taskAPI.GET("/:taskID", handler.GetTaskStatusHandler)
		taskAPI.GET("/run_again/:taskID", handler.GetTaskStatusHandler)
	}

	return r
}

func Run(port int) {
	Get().Run(fmt.Sprintf(":%d", port))
}
