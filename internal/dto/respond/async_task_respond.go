package respond

// AsyncTaskRespond 异步任务响应结构体
type AsyncTaskRespond struct {
	TaskType string      `json:"task_type"`  // 任务类型
	TaskId   string      `json:"task_id"`    // 任务ID
	Success  bool        `json:"success"`    // 是否成功
	Message  string      `json:"message"`    // 响应消息
	Data     interface{} `json:"data"`       // 响应数据
}

// AsyncLoadingRespond 异步加载中的响应
type AsyncLoadingRespond struct {
	Loading bool   `json:"loading"` // 是否正在加载
	TaskId  string `json:"task_id"` // 任务ID
	Message string `json:"message"` // 提示消息
}