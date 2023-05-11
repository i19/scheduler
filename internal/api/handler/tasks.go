package handler

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"scheduler/internal/config"
	"scheduler/internal/platform/database"
	"scheduler/pkg/errors"
	"scheduler/pkg/scheduler"
	"scheduler/pkg/task"
	"scheduler/pkg/utils"
)

const taskIDLength = 32

func submitTaskHandler(c *gin.Context, req *SubmitTaskRequest) {
	if !config.HasKind(req.Kind) || !config.HasPriority(req.Priority) {
		Response(c, errors.BadRequest, fmt.Sprintf("task kind %s or priority %s is not allowed", req.Kind, req.Priority), nil)
		return
	}

	taskByte, err := json.Marshal(req.Task)
	if err != nil {
		Response(c, errors.BadRequest, "task field should be a json", nil)
		return
	}

	taskID := utils.RandStringRunes(taskIDLength)
	taskObject, err := task.CreateTask(taskID, req.Kind, req.Priority, string(taskByte))
	if err != nil {
		Response(c, errors.InternalServerError, fmt.Sprintf("create task error: %s", err.Error()), nil)
		return
	}

	go scheduler.SubmitTask(taskObject)
	ResponseOK(c, &SubmitTaskResponse{TaskID: taskObject.TaskID})
}

type SubmitTaskRequest struct {
	Kind     string      `json:"kind" binding:"required"`
	Priority string      `json:"priority" binding:"required,oneof=high medium low"`
	Task     interface{} `json:"task" binding:"required"`
	//Tasks []*SubTaskRequest `json:"tasks" binding:"required"`
}
type SubmitTaskResponse struct {
	TaskID string `json:"task_id"`
}

// todo 任务关系
//type SubTaskRequest struct {
//	Type         string            `json:"type" binding:"required"`
//	Priorities     string            `json:"priority" binding:"required,oneof=high medium low"`
//	Dependencies []*SubTaskRequest `json:"dependencies"`
//	Task         interface{}       `json:"data" binding:"required"`
//}

// SubmitTaskHandler
// @Summary 提交任务
// @Description 提交任务到Scheduler
// @Tags Tasks
// @Accept json
// @Produce json
// @Param   task body SubmitTaskRequest true "任务对象"
// @Success 200 {object} handler.ResponseProtocol{}
// @Failure 200 {object} handler.ResponseProtocol{}
// @Router /api/tasks [post]
func SubmitTaskHandler(c *gin.Context) {
	var req SubmitTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Response(c, errors.BadRequest, fmt.Sprintf("parse request error: %s", err.Error()), nil)
		return
	}

	submitTaskHandler(c, &req)
}

// GetTaskStatusHandler
// @Summary 查询任务状态
// @Description 查询特定任务的执行状态
// @Tags Tasks
// @Produce json
// @Param taskID path string true "任务ID"
// @Success 200 {object} handler.ResponseProtocol{}
// @Failure 200 {object} handler.ResponseProtocol{}
// @Router /api/tasks/{taskID} [get]
func GetTaskStatusHandler(c *gin.Context) {
	taskID := c.Query("taskID")
	if len(taskID) != taskIDLength {
		Response(c, errors.BadRequest, "invalid task id", nil)
		return
	}

	var x task.Task
	if err := database.Engine.Where(&task.Task{TaskID: taskID}).First(&x).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Response(c, errors.ResourceNotExist, "task id not exist", nil)
			return
		} else {
			Response(c, errors.InternalServerError, err.Error(), nil)
			return
		}
	}
	Response(c, errors.OK, "", &x)
}

// GetTaskListHandler
// @Summary 查询任务列表
// @Description 查询任务列表，可根据状态、类型和优先级过滤
// @Tags Tasks
// @Produce json
// @Param status query string false "任务状态"
// @Param type query string false "任务类型"
// @Param priority query string false "任务优先级"
// @Success 200 {object} handler.ResponseProtocol{}
// @Failure 200 {object} handler.ResponseProtocol{}
// @Router /api/tasks [get]
// todo 分页
func GetTaskListHandler(c *gin.Context) {
	var x []*task.Task
	if err := database.Engine.Find(&x).Error; err != nil {
		Response(c, errors.InternalServerError, err.Error(), nil)
		return
	}

	Response(c, errors.OK, "", x)
}

// RunAgainHandler
// @Summary 重跑任务
// @Description 将指定任务再次运行，此过程会返回新的 taskID
// @Tags Tasks
// @Accept json
// @Produce json
// @Param taskID path string true "任务ID"
// @Success 200 {object} handler.ResponseProtocol{}
// @Failure 200 {object} handler.ResponseProtocol{}
// @Router /api/tasks/run_again/{taskID} [post]
func RunAgainHandler(c *gin.Context) {
	taskID := c.Query("taskID")
	if len(taskID) != taskIDLength {
		Response(c, errors.BadRequest, "invalid task id", nil)
		return
	}

	var x task.Task
	if err := database.Engine.Where(&task.Task{TaskID: taskID}).First(&x).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Response(c, errors.ResourceNotExist, "task id not exist", nil)
			return
		} else {
			Response(c, errors.InternalServerError, err.Error(), nil)
			return
		}
	}
	var y interface{}
	if err := json.Unmarshal([]byte(x.Task), &y); err != nil {
		Response(c, errors.InternalServerError, err.Error(), nil)
		return
	}

	submitTaskHandler(c, &SubmitTaskRequest{
		Kind:     x.Kind,
		Priority: x.Priority,
		Task:     nil,
	})
}
