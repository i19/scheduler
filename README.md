## 介绍

scheduler 是一个任务队列调度服务，具备以下特性

1. 提供 HTTP 结构实现任务管理（提交、查看、重新运行）
2. 已正常提交的任务不会丢失，通过 MySQL 管理，服务重启会自动加载并继续调度
3. 可通过 etcd 自动发现 worker 节点的增删，能感知 worker 节点状态，异常节点状态恢复后会继续提供服务

## 设计

### 接口

返回为统一 json 格式

```json
{
  "code": 0,   // code 为 0 则为成功，其他为异常，有 error code 对照，参看 pkg/errors/error.go
  "message": "hello world", // 提示信息
  "data": {json object}, // 如果需要返回数据，会放在该处
}
```


#### 任务提交

POST /api/tasks

Request

| 参数       | 数据类型        | 是否必须 | 含义                      | 
|----------|-------------|------|-------------------------|
| kind     | string（预定义） | Y    | 任务类型                    |
| priority | string（预定义） | Y    | 优先级                     |   
| task     | json 对象     | Y    | 任务参数（该对象会被原样转交给 worker） |

Response 
```json
{
  "task_id": "xxxx", // 长度为 32
}
```

Todo
1. 任务关系（依赖，并行等...）

#### 查看任务列表
GET /api/tasks
TODO: 分页

Response
```json
[
  {
      "task_id": "xxxx",
      "kind": "xxx",
      "priority": "high",
      "task": json object,
      
      // 下面是运行相关信息
      "worker_address": "1.1.1.1" // 运行该任务的节点，如果为空，意味着该任务在等待运行中
      "start_time": 1683812180, // 开始运行的时间戳
      "end_time": 1683812190, // 任务运行结束的时间戳
      "is_succeed": true, // 任务运行是否成功（当 end_time >0 时，该值才有意义）
      "message": "hello world", // worker 运行完成后给出的提示信息（异常？结果提示？）
  }
]
```

#### 查看单个任务详情

GET /api/tasks/:taskID

Response
```json
{
  "task_id": "xxxx",
  "kind": "xxx",
  "priority": "high",
  "task": json object,
  
  // 下面是运行相关信息
  "worker_address": "1.1.1.1" // 运行该任务的节点，如果为空，意味着该任务在等待运行中
  "start_time": 1683812180, // 开始运行的时间戳
  "end_time": 1683812190, // 任务运行结束的时间戳
  "is_succeed": true, // 任务运行是否成功（当 end_time >0 时，该值才有意义）
  "message": "hello world", // worker 运行完成后给出的提示信息（异常？结果提示？）
}
```

#### 重新运行某一个任务

POST /api/tasks/run_again/:taskID

Response
```json
{
  "task_id": "xxxx", // 长度为 32
}
```

TODO
#### 任务停止
#### 状态统计、监控 etc

### 详细设计
#### 启动过程
1. 一些外部资源的初始化（数据库, etcd）
2. 通过 etcd 获取对 worker，并按照 任务类型 进行分组，并更新其状态，如果空闲则将其加入 worker 待调度队列，否则在后台进行监控等待其恢复空闲状态，再加入 worker 待调度队列 
3. 将预定义好的 任务类型以及优先级 进行加载，按照任务类型生成多个 任务优先级队列（配置可以考虑动态化）
4. 启动 scheduler 模块，按照 任务类型，将 任务优先级队列 与 worker 待调度队列进行绑定
5. 从数据库中加载已提交但是尚未被分发的任务，将其加载进 任务优先级队列中
6. 启动 API 对外提供服务

#### 提交任务的数据流程
1. 任务通过 API 提交后，进行合法性检查
2. assign 新的 `taskID`，并将其存放在数据库中，如果异常，则通过接口返回进行报错
3. 通过 `scheduler` 将其加载进任务队列中，等待调度
4. 当其按照 优先级规则 被调度出来以后，等待空闲 `worker`
5. 当有 `worker` 可以提供服务时，将该任务进行分发
6. 后台启动 `worker.waitToFreeOrStop` 等待 worker 运行结束，并获取结果
7. 将结果同步到数据库中
8. 如果 `worker` 状态正常，且未退出（感知 etcd 状态） ，将该 worker 重新加入 worker 待调度队列，等待下次调度

#### 优先级调度规则
目前是指数级实现，曲线可通过配置文件中的中的 `priorityfactor` 进行调整
```
priorityfactor: = 2 时，各优先级的输出比例为 high : mid : low = 4 : 2 : 1  
假设每个优先级的任务都有 10 个时，输出如下，可通过  pkg/queue/queue_test.go 测试
high_1,high_2,high_3,high_4,mid_1,mid_2,low_1
high_5,high_6,high_7,high_8,mid_3,mid_4,low_2,
high_9,mid_5,mid_6,low_3,
mid_7,mid_8,low_4,
mid_9,low_5,
low_6,low_7,low_8,low_9

priorityfactor: = 10 , 各优先级的输出比例为 high : mid : low = 100 : 10 : 1
```