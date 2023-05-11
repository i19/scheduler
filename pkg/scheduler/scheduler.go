package scheduler

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"scheduler/internal/config"
	"scheduler/internal/platform/database"
	"scheduler/internal/platform/worker"
	"scheduler/pkg/queue"
	"scheduler/pkg/task"
)

var (
	// 该 map 仅在初始化时有写入，所以不使用 lock 保护
	queueByTaskKind = make(map[string]*queue.PriorityQueue)
	once            sync.Once
)

// addKind 为每一个类型创建 任务优先级队列，以及任务调度和结果处理资源
func addKind(kind string, priorities []string) {
	if _, ok := queueByTaskKind[kind]; !ok {
		pq := queue.NewPriorityQueue(priorities, config.Config.PriorityFactor)
		queueByTaskKind[kind] = pq

		//resultChannel  任务异步执行结果同步到此处
		resultChannel := make(chan *worker.TaskReport, 1e6)

		// 消费 resultChannel ， 将结果同步到数据库中
		go func() {
			//todo 状态暂存机制
			for result := range resultChannel {
				if err := database.Engine.Model(&task.Task{TaskID: result.TaskID}).Updates(task.Task{
					WorkerAddress: result.WorkerAddress,
					StartTime:     result.StartTime,
					EndTime:       result.EndTime,
					IsSucceed:     result.IsSucceed,
					Message:       result.Message,
				}).Error; err != nil {
					logrus.Errorf("failed to save result for task %s due to database error: %s", result.TaskID, err.Error())
				}
			}
		}()

		// 消费任务优先级队列，进行调度
		go func() {
			for ele := range pq.Pop() {
				submitted := false
				submitTimes := 1
				for !submitted {
					t := ele.(*task.Task)

					// 如果没有空闲 worker，此步会卡住等待
					if workerAddress, submitError := worker.RunTask(t, resultChannel); submitError != nil {
						logrus.Errorf("submit task %s %d times error, will retry: %s", t.TaskID, submitTimes, submitError.Error())
						submitTimes++
					} else {
						submitted = true
						logrus.Errorf("submit task %s to %s", t.TaskID, workerAddress)
						t.WorkerAddress = workerAddress
						t.StartTime = time.Now().Unix()
						if err := database.Engine.Save(t).Error; err != nil {
							// 该状态可以通过 result 处理 cover，仅仅是为了快速更新
							logrus.Errorf("save task %s status error :%s", t.TaskID, err.Error())
						}
					}
				}
			}
		}()
	}
}

// Init
// 获取当前所有 kind 并进行初始化
// 将之前尚未运行的 task 进行加载
func Init(preDefinedKinds []string, preDefinedPriorities []string) {
	once.Do(func() {
		for _, kind := range preDefinedKinds {
			addKind(kind, preDefinedPriorities)
		}

		tasks, err := task.GetAllUnDispatchedTask()
		if err != nil {
			logrus.Panicf("scheduler init failed to load undispatched task: %s", err.Error())
		}
		for _, task := range tasks {
			if err := task.Parse(); err != nil {
				logrus.Errorf("skip task %s due to: %s", task.TaskID, err.Error())
			} else {
				SubmitTask(task)
			}
		}
	})
}

func SubmitTask(task *task.Task) {
	queueByTaskKind[task.Kind].Push(task.Priority, task)
}
