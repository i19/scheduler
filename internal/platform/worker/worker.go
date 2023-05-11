package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"scheduler/internal/platform/etcd"
	"scheduler/pkg/task"
)

var (
	workers          = make(map[string]map[string]*worker)
	freeWorkerByKind = make(map[string]chan *worker)
	once             sync.Once
)

func Init(workerNodesPath string) {
	once.Do(
		func() {
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			resp, err := etcd.Client.Get(ctx, workerNodesPath)
			if err != nil {
				panic(fmt.Sprintf("Failed to get all lintingRule from etcd: %s", err.Error()))
			}

			workerCount := 0
			for _, kv := range resp.Kvs {
				// example "/a/b/c/d/e/task_download/1.1.1.1:80
				// example "/a/b/c/d/e/task_upload/2.2.2.2:80
				split := strings.Split(string(kv.Key), "/")
				kind, address := split[len(split)-2], split[len(split)-1]
				addWorker(kind, address)
				workerCount++
			}

			for kind, workersByKind := range workers {
				for _, x := range workersByKind {
					go func(w *worker) {
						w.waitToFreeOrStop()
						if w.isFree() {
							freeWorkerByKind[kind] <- w
						}
					}(x)
				}
			}

			go func() {
				for change := range etcd.Client.Watch(context.Background(), workerNodesPath) {
					for _, event := range change.Events {
						switch event.Type.String() {
						case etcd.ActionPUT:
							split := strings.Split(string(event.Kv.Key), "/")
							kind, address := split[len(split)-2], split[len(split)-1]
							addWorker(kind, address)
						case etcd.ActionDELETE:
							split := strings.Split(string(event.Kv.Key), "/")
							kind, address := split[len(split)-2], split[len(split)-1]
							delWorker(kind, address)
						}
					}
				}
			}()
		},
	)
}

func addWorker(kind, address string) {
	if _, ok := workers[kind]; !ok {
		workers[kind] = make(map[string]*worker)
		freeWorkerByKind[kind] = make(chan *worker)
	}
	if _, ok := workers[kind][address]; !ok {
		x := &worker{
			kind:    kind,
			address: address,
		}
		workers[kind][address] = x

	}
}

func delWorker(kind, address string) {
	if _, ok := workers[kind]; ok {
		if x, ok := workers[kind][address]; ok {
			x.stop()
			delete(workers[kind], address)
		}
	}
}

type TaskReport struct {
	Kind          string
	TaskID        string
	StartTime     int64
	EndTime       int64
	WorkerAddress string
	IsSucceed     bool
	Message       string
}

func RunTask(taskObj *task.Task, result chan *TaskReport) (string, error) {
	w := <-freeWorkerByKind[taskObj.Kind]
	if !w.isFree() {
		return "", fmt.Errorf("this node is not ok")
	}

	if err := w.deliver(taskObj.TaskID, taskObj.TaskObject); err != nil {
		// 该节点可能故障，需要处理，不需要占用前台
		go func() {
			// 等到该节点网络可达，或者该节点被明确删除。 如果可用才将该节点放队列等待下次任务调用
			w.waitToFreeOrStop()
			if w.isFree() {
				freeWorkerByKind[taskObj.Kind] <- w
			}
		}()
		return "", err
	} else {
		// 后台等待 task 任务完成以后，再将该节点放回队列
		go func() {
			w.waitToFreeOrStop()
			if w.isFree() {
				freeWorkerByKind[taskObj.Kind] <- w
			}
			result <- w.reportTask()
		}()
		return w.address, nil
	}
}
