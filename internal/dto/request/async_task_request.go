package request

// AsyncTaskRequest 异步任务请求结构体
type AsyncTaskRequest struct {
	TaskType   string      `json:"task_type"`   // 任务类型：load_message_list, load_group_message_list, load_joined_group_list
	TaskId     string      `json:"task_id"`     // 任务ID，用于追踪
	ClientId   string      `json:"client_id"`   // WebSocket客户端ID
	UserId     string      `json:"user_id"`     // 用户ID
	Parameters interface{} `json:"parameters"`  // 任务参数
}

// MessageListTaskParams 聊天记录加载任务参数
type MessageListTaskParams struct {
	UserOneId string `json:"user_one_id"`
	UserTwoId string `json:"user_two_id"`
}

// GroupMessageListTaskParams 群聊记录加载任务参数
type GroupMessageListTaskParams struct {
	GroupId string `json:"group_id"`
}

// JoinedGroupListTaskParams 加入群聊列表加载任务参数
type JoinedGroupListTaskParams struct {
	OwnerId string `json:"owner_id"`
}