package gorm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"io"
	"kama_chat_server/internal/config"
	"kama_chat_server/internal/dao"
	"kama_chat_server/internal/dto/request"
	"kama_chat_server/internal/dto/respond"
	"kama_chat_server/internal/model"
	myKafka "kama_chat_server/internal/service/kafka"
	myredis "kama_chat_server/internal/service/redis"
	"kama_chat_server/pkg/constants"
	"kama_chat_server/pkg/util/random"
	"kama_chat_server/pkg/zlog"
	"os"
	"path/filepath"
)

type messageService struct {
}

var MessageService = new(messageService)

// GetMessageList 获取聊天记录
func (m *messageService) GetMessageList(userOneId, userTwoId string) (string, interface{}, int) {
	// 检查是否启用异步模式
	kafkaConfig := config.GetConfig().KafkaConfig
	if kafkaConfig.MessageMode == "kafka" || kafkaConfig.MessageMode == "hybrid" {
		return m.asyncLoadMessageList(userOneId, userTwoId)
	}
	return m.syncLoadMessageList(userOneId, userTwoId)
}

// asyncLoadMessageList 异步加载聊天记录
func (m *messageService) asyncLoadMessageList(userOneId, userTwoId string) (string, interface{}, int) {
	// 生成任务ID
	taskId := fmt.Sprintf("ML%s", random.GetNowAndLenRandomString(11))
	
	// 创建异步任务
	taskParams := request.MessageListTaskParams{
		UserOneId: userOneId,
		UserTwoId: userTwoId,
	}
	
	asyncTask := request.AsyncTaskRequest{
		TaskType:   "load_message_list",
		TaskId:     taskId,
		ClientId:   userOneId, // 使用userOneId作为客户端ID
		UserId:     userOneId,
		Parameters: taskParams,
	}
	
	// 发送任务到Kafka
	taskData, err := json.Marshal(asyncTask)
	if err != nil {
		zlog.Error("序列化异步任务失败: " + err.Error())
		return m.syncLoadMessageList(userOneId, userTwoId) // 降级到同步处理
	}
	
	err = myKafka.KafkaService.AsyncTaskWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(taskId),
		Value: taskData,
	})
	
	if err != nil {
		zlog.Error("发送异步任务到Kafka失败: " + err.Error())
		return m.syncLoadMessageList(userOneId, userTwoId) // 降级到同步处理
	}
	
	// 返回加载中状态
	loadingResp := respond.AsyncLoadingRespond{
		Loading: true,
		TaskId:  taskId,
		Message: "正在加载聊天记录...",
	}
	
	return "正在加载聊天记录...", loadingResp, 200
}

// syncLoadMessageList 同步加载聊天记录
func (m *messageService) syncLoadMessageList(userOneId, userTwoId string) (string, interface{}, int) {
	rspString, err := myredis.GetKeyNilIsErr("message_list_" + userOneId + "_" + userTwoId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			zlog.Info(err.Error())
			zlog.Info(fmt.Sprintf("%s %s", userTwoId, userTwoId))
			var messageList []model.Message
			if res := dao.GormDB.Where("(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)", userOneId, userTwoId, userTwoId, userOneId).Order("created_at ASC").Find(&messageList); res.Error != nil {
				zlog.Error(res.Error.Error())
				return constants.SYSTEM_ERROR, nil, -1
			}
			var rspList []respond.GetMessageListRespond
			for _, message := range messageList {
				rspList = append(rspList, respond.GetMessageListRespond{
					SendId:     message.SendId,
					SendName:   message.SendName,
					SendAvatar: message.SendAvatar,
					ReceiveId:  message.ReceiveId,
					Content:    message.Content,
					Url:        message.Url,
					Type:       message.Type,
					FileType:   message.FileType,
					FileName:   message.FileName,
					FileSize:   message.FileSize,
					CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
			//rspString, err := json.Marshal(rspList)
			//if err != nil {
			//	zlog.Error(err.Error())
			//}
			//if err := myredis.SetKeyEx("message_list_"+userOneId+"_"+userTwoId, string(rspString), time.Minute*constants.REDIS_TIMEOUT); err != nil {
			//	zlog.Error(err.Error())
			//}
			return "获取聊天记录成功", rspList, 0
		} else {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	var rsp []respond.GetMessageListRespond
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		zlog.Error(err.Error())
	}
	return "获取群聊记录成功", rsp, 0
}

// GetGroupMessageList 获取群聊消息记录
func (m *messageService) GetGroupMessageList(groupId string) (string, interface{}, int) {
	// 检查是否启用异步模式
	kafkaConfig := config.GetConfig().KafkaConfig
	if kafkaConfig.MessageMode == "kafka" || kafkaConfig.MessageMode == "hybrid" {
		return m.asyncLoadGroupMessageList(groupId)
	}
	return m.syncLoadGroupMessageList(groupId)
}

// asyncLoadGroupMessageList 异步加载群聊消息记录
func (m *messageService) asyncLoadGroupMessageList(groupId string) (string, interface{}, int) {
	// 生成任务ID
	taskId := fmt.Sprintf("GML%s", random.GetNowAndLenRandomString(11))
	
	// 创建异步任务
	taskParams := request.GroupMessageListTaskParams{
		GroupId: groupId,
	}
	
	asyncTask := request.AsyncTaskRequest{
		TaskType:   "load_group_message_list",
		TaskId:     taskId,
		ClientId:   groupId, // 使用groupId作为客户端ID
		UserId:     groupId,
		Parameters: taskParams,
	}
	
	// 发送任务到Kafka
	taskData, err := json.Marshal(asyncTask)
	if err != nil {
		zlog.Error("序列化异步任务失败: " + err.Error())
		return m.syncLoadGroupMessageList(groupId) // 降级到同步处理
	}
	
	err = myKafka.KafkaService.AsyncTaskWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(taskId),
		Value: taskData,
	})
	
	if err != nil {
		zlog.Error("发送异步任务到Kafka失败: " + err.Error())
		return m.syncLoadGroupMessageList(groupId) // 降级到同步处理
	}
	
	// 返回加载中状态
	loadingResp := respond.AsyncLoadingRespond{
		Loading: true,
		TaskId:  taskId,
		Message: "正在加载群聊记录...",
	}
	
	return "正在加载群聊记录...", loadingResp, 200
}

// syncLoadGroupMessageList 同步加载群聊消息记录
func (m *messageService) syncLoadGroupMessageList(groupId string) (string, interface{}, int) {
	rspString, err := myredis.GetKeyNilIsErr("group_messagelist_" + groupId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			var messageList []model.Message
			if res := dao.GormDB.Where("receive_id = ?", groupId).Order("created_at ASC").Find(&messageList); res.Error != nil {
				zlog.Error(res.Error.Error())
				return constants.SYSTEM_ERROR, nil, -1
			}
			var rspList []respond.GetGroupMessageListRespond
			for _, message := range messageList {
				rsp := respond.GetGroupMessageListRespond{
					SendId:     message.SendId,
					SendName:   message.SendName,
					SendAvatar: message.SendAvatar,
					ReceiveId:  message.ReceiveId,
					Content:    message.Content,
					Url:        message.Url,
					Type:       message.Type,
					FileType:   message.FileType,
					FileName:   message.FileName,
					FileSize:   message.FileSize,
					CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
				}
				rspList = append(rspList, rsp)
			}
			//rspString, err := json.Marshal(rspList)
			//if err != nil {
			//	zlog.Error(err.Error())
			//}
			//if err := myredis.SetKeyEx("group_messagelist_"+groupId, string(rspString), time.Minute*constants.REDIS_TIMEOUT); err != nil {
			//	zlog.Error(err.Error())
			//}
			return "获取聊天记录成功", rspList, 0
		} else {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	var rsp []respond.GetGroupMessageListRespond
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		zlog.Error(err.Error())
	}
	return "获取聊天记录成功", rsp, 0
}

// UploadAvatar 上传头像
func (m *messageService) UploadAvatar(c *gin.Context) (string, int) {
	if err := c.Request.ParseMultipartForm(constants.FILE_MAX_SIZE); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	mForm := c.Request.MultipartForm
	for key, _ := range mForm.File {
		file, fileHeader, err := c.Request.FormFile(key)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer file.Close()
		zlog.Info(fmt.Sprintf("文件名：%s，文件大小：%d", fileHeader.Filename, fileHeader.Size))
		// 原来Filename应该是213451545.xxx，将Filename修改为avatar_ownerId.xxx
		ext := filepath.Ext(fileHeader.Filename)
		zlog.Info(ext)
		localFileName := config.GetConfig().StaticAvatarPath + "/" + fileHeader.Filename
		out, err := os.Create(localFileName)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		zlog.Info("完成文件上传")
	}
	return "上传成功", 0
}

// UploadFile 上传文件
func (m *messageService) UploadFile(c *gin.Context) (string, int) {
	if err := c.Request.ParseMultipartForm(constants.FILE_MAX_SIZE); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	mForm := c.Request.MultipartForm
	for key, _ := range mForm.File {
		file, fileHeader, err := c.Request.FormFile(key)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer file.Close()
		zlog.Info(fmt.Sprintf("文件名：%s，文件大小：%d", fileHeader.Filename, fileHeader.Size))
		// 原来Filename应该是213451545.xxx，将Filename修改为avatar_ownerId.xxx
		ext := filepath.Ext(fileHeader.Filename)
		zlog.Info(ext)
		localFileName := config.GetConfig().StaticFilePath + "/" + fileHeader.Filename
		out, err := os.Create(localFileName)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		zlog.Info("完成文件上传")
	}
	return "上传成功", 0
}
