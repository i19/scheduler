package task

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"scheduler/internal/config"
	"scheduler/internal/platform/database"
)

type Task struct {
	// 任务创建
	gorm.Model
	TaskID   string `gorm:"type:char(64);UNIQUE_INDEX;not null" json:"task_id"`
	Kind     string `gorm:"type:varchar(32);not null" json:"kind"`
	Task     string `gorm:"type:mediumtext;not null" json:"-"`
	Priority string `gorm:"type:varchar(32);not null" json:"priority"`

	// 任务分配
	WorkerAddress string `gorm:"type:varchar(64);" json:"worker_address"`
	StartTime     int64  `gorm:"type:int;not null; default: 0" json:"start_time"`

	// 任务完成
	EndTime   int64  `gorm:"type:int;not null; default: 0" json:"end_time"`
	IsSucceed bool   `gorm:"default: false" json:"is_succeed"`
	Message   string `gorm:"type:mediumtext" json:"message"`

	TaskObject interface{} `gorm:"-" json:"task"`
}

func CreateTask(taskID, kind, priority string, taskString string) (*Task, error) {
	if err := database.Engine.Create(&Task{
		TaskID:     taskID,
		Kind:       kind,
		Task:       taskString,
		Priority:   priority,
		TaskObject: taskString,
	}).Error; err != nil {
		return nil, err
	}
	return nil, nil
}

// GetAllUnDispatchedTask 获取已提交但尚未运行的 task
// todo 分页加载
func GetAllUnDispatchedTask() ([]*Task, error) {
	var result []*Task
	if err := database.Engine.Where(&Task{WorkerAddress: ""}).Find(&result).Error; err != nil {
		return nil, err
	}
	for _, x := range result {
		if err := json.Unmarshal([]byte(x.Task), &x.TaskObject); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (t *Task) Parse() error {
	if !config.HasKind(t.Kind) || !config.HasPriority(t.Priority) {
		return fmt.Errorf("task %s kind %s or priority %s is not allowed", t.TaskID, t.Kind, t.Priority)
	}

	if err := json.Unmarshal([]byte(t.Task), &t.TaskObject); err != nil {
		return fmt.Errorf("task %s task format error: %s", t.TaskID, err.Error())
	}

	return nil
}

//func GetAllUnFinishedTask() ([]*Task, error) {
//	var result []*Task
//	if err := database.Engine.Where(&Task{EndTime: 0}).Not(&Task{StartTime: 0}).Find(&result).Error; err != nil {
//		return nil, err
//	}
//	return result, nil
//}
