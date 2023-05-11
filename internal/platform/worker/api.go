package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

type workerStatus struct {
	Kind string `json:"kind"`
	// TaskID 最后一个任务 ID。 如果有正在执行的任务，则为当前任务 ID 否则为上次的任务 ID 。 如果为空， 则表示该节点尚未运行过任务
	TaskID string `json:"task_id"`
	// 任务的执行状态，如果 TaskID 字段为空，那么该字段无意义。如果 TaskID 不为空，1 代表运行中， 2 代表运行无异常完成， 3 代表有运行有异常完成
	Code int `json:"code"`
	// task 运行提示信息（异常信息，该信息通过页面可以查看）
	Message string `json:"message"`
}

func (s *workerStatus) isFree() bool {
	return s.TaskID == "" || s.Code == 2 || s.Code == 3
}

type deliverTaskRequest struct {
	TaskID string
	Task   interface{}
}
type deliverTaskResponse struct {
	TaskID       string
	IsOK         bool
	ErrorMessage string
}

type worker struct {
	kind    string
	address string
	// 节点可以连通
	isOK bool
	// 节点是否停止服务，如果是，需要停止后台状态更新
	isStop bool
	lock   sync.Mutex

	// statusUpdateTime 晚于 taskDeliveredTime 时， status 才有意义
	statusUpdateTime  time.Time
	taskDeliveredTime time.Time
	status            workerStatus
}

func (w *worker) stop() {
	w.isStop = true
}

func (w *worker) isFree() bool {
	if !w.isStop && w.isOK && w.statusUpdateTime.After(w.taskDeliveredTime) && w.status.isFree() {
		return true
	}
	return false
}

func (w *worker) deliver(taskID string, task interface{}) error {
	_, body, errs := gorequest.New().
		Post(fmt.Sprintf("http://%s/submit_task", w.address)).
		SendStruct(deliverTaskRequest{
			TaskID: taskID,
			Task:   task,
		}).
		Timeout(time.Second*3).
		Retry(3, time.Second*5, http.StatusRequestTimeout, http.StatusGatewayTimeout, http.StatusInternalServerError).
		End()
	if len(errs) != 0 {
		return errs[0]
	}

	var resp deliverTaskResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return fmt.Errorf("deliver task to %s response invalid format: %s", w.address, err.Error())
	}

	if resp.TaskID != taskID && !resp.IsOK {
		return fmt.Errorf("deliver task to %s response error: %s", w.address, resp.ErrorMessage)
	}

	w.taskDeliveredTime = time.Now()
	return nil
}
func (w *worker) updateStatus() {
	now := time.Now()
	_, body, errs := gorequest.New().
		Get(fmt.Sprintf("http://%s/status", w.address)).
		Timeout(time.Second*3).
		Retry(3, time.Second*5, http.StatusRequestTimeout, http.StatusGatewayTimeout, http.StatusInternalServerError).
		End()

	if len(errs) != 0 {
		w.isOK = false
		logrus.Errorf("worker node %s-%s return error: %s", w.kind, w.address, errs[0].Error())
		time.Sleep(time.Second * 30)
		return
	}
	if err := json.Unmarshal([]byte(body), &w.status); err != nil {
		w.isOK = false
		logrus.Errorf("worker node %s-%s return invalid format: %s", w.kind, w.address, err.Error())
		time.Sleep(time.Second * 30)
		return
	}

	w.statusUpdateTime = now
}

// todo 超时处理
func (w *worker) waitToFreeOrStop() {
	for {
		if w.isStop {
			return
		}
		w.updateStatus()
		if w.isFree() {
			return
		}
	}
}

func (w *worker) reportTask() *TaskReport {
	return &TaskReport{
		Kind:          w.kind,
		TaskID:        w.status.TaskID,
		StartTime:     w.taskDeliveredTime.Unix(),
		EndTime:       w.statusUpdateTime.Unix(),
		IsSucceed:     w.status.Code == 2,
		Message:       w.status.Message,
		WorkerAddress: w.address,
	}
}
