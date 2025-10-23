package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"kama_chat_server/internal/config"
	"kama_chat_server/internal/dao"
	"kama_chat_server/internal/dto/request"
	"kama_chat_server/internal/dto/respond"
	"kama_chat_server/internal/model"
	myKafka "kama_chat_server/internal/service/kafka"
	myredis "kama_chat_server/internal/service/redis"
	"kama_chat_server/pkg/constants"
	"kama_chat_server/pkg/enum/message/message_status_enum"
	"kama_chat_server/pkg/enum/message/message_type_enum"
	"kama_chat_server/pkg/util/random"
	"kama_chat_server/pkg/zlog"
	"strconv"
	"sync"
	"time"
)

// HybridServer 混合服务器，支持channel和kafka动态切换
type HybridServer struct {
	Clients         map[string]*Client
	mutex           *sync.Mutex
	Transmit        chan []byte  // channel模式的转发通道
	Login           chan *Client // 登录通道
	Logout          chan *Client // 退出登录通道
	
	// 监控相关
	channelMonitor  *ChannelMonitor
	useKafka        bool // 当前是否使用kafka模式
	modeMutex       *sync.RWMutex
}

// ChannelMonitor channel监控器
type ChannelMonitor struct {
	thresholdRatio    float64       // 阈值比例
	checkInterval     time.Duration // 检查间隔
	overloadDuration  time.Duration // 持续超载时间阈值
	overloadStartTime *time.Time    // 开始超载的时间
	isOverloaded      bool          // 当前是否超载
	mutex             *sync.RWMutex
}

var HybridChatServer *HybridServer

func init() {
	if HybridChatServer == nil {
		kafkaConfig := config.GetConfig().KafkaConfig
		
		HybridChatServer = &HybridServer{
			Clients:  make(map[string]*Client),
			mutex:    &sync.Mutex{},
			Transmit: make(chan []byte, constants.CHANNEL_SIZE),
			Login:    make(chan *Client, constants.CHANNEL_SIZE),
			Logout:   make(chan *Client, constants.CHANNEL_SIZE),
			
			channelMonitor: &ChannelMonitor{
				thresholdRatio:   kafkaConfig.HybridThreshold,
				checkInterval:    time.Duration(kafkaConfig.HybridMonitorInterval) * time.Second,
				overloadDuration: time.Duration(kafkaConfig.HybridSwitchDuration) * time.Second,
				mutex:           &sync.RWMutex{},
			},
			useKafka:  false,
			modeMutex: &sync.RWMutex{},
		}
	}
}

// Start 启动混合服务器
func (h *HybridServer) Start() {
	defer func() {
		if r := recover(); r != nil {
			zlog.Error(fmt.Sprintf("hybrid server panic: %v", r))
		}
		close(h.Transmit)
		close(h.Login)
		close(h.Logout)
	}()

	// 启动channel监控
	go h.startChannelMonitor()
	
	// 启动kafka消息读取（如果配置了kafka）
	kafkaConfig := config.GetConfig().KafkaConfig
	if kafkaConfig.MessageMode == "hybrid" || kafkaConfig.MessageMode == "kafka" {
		go h.startKafkaMessageReader()
	}

	// 主消息处理循环
	for {
		select {
		case client := <-h.Login:
			{
				h.mutex.Lock()
				h.Clients[client.Uuid] = client
				h.mutex.Unlock()
				zlog.Debug(fmt.Sprintf("欢迎来到kama聊天服务器，亲爱的用户%s\n", client.Uuid))
				err := client.Conn.WriteMessage(websocket.TextMessage, []byte("欢迎来到kama聊天服务器"))
				if err != nil {
					zlog.Error(err.Error())
				}
			}

		case client := <-h.Logout:
			{
				h.mutex.Lock()
				delete(h.Clients, client.Uuid)
				h.mutex.Unlock()
				zlog.Info(fmt.Sprintf("用户%s退出登录\n", client.Uuid))
				if err := client.Conn.WriteMessage(websocket.TextMessage, []byte("已退出登录")); err != nil {
					zlog.Error(err.Error())
				}
			}

		case data := <-h.Transmit:
			{
				// 处理channel模式的消息
				h.processMessage(data)
			}
		}
	}
}

// startChannelMonitor 启动channel监控
func (h *HybridServer) startChannelMonitor() {
	ticker := time.NewTicker(h.channelMonitor.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		h.checkChannelLoad()
	}
}

// checkChannelLoad 检查channel负载
func (h *HybridServer) checkChannelLoad() {
	currentLoad := len(h.Transmit)
	threshold := int(float64(constants.CHANNEL_SIZE) * h.channelMonitor.thresholdRatio)
	
	h.channelMonitor.mutex.Lock()
	defer h.channelMonitor.mutex.Unlock()
	
	if currentLoad >= threshold {
		// 超过阈值
		if !h.channelMonitor.isOverloaded {
			// 刚开始超载
			now := time.Now()
			h.channelMonitor.overloadStartTime = &now
			h.channelMonitor.isOverloaded = true
			zlog.Info(fmt.Sprintf("Channel开始超载: %d/%d (阈值: %d)", currentLoad, constants.CHANNEL_SIZE, threshold))
		} else {
			// 持续超载，检查是否需要切换到kafka
			if time.Since(*h.channelMonitor.overloadStartTime) >= h.channelMonitor.overloadDuration {
				h.switchToKafka()
			}
		}
	} else {
		// 负载正常
		if h.channelMonitor.isOverloaded {
			h.channelMonitor.isOverloaded = false
			h.channelMonitor.overloadStartTime = nil
			zlog.Info(fmt.Sprintf("Channel负载恢复正常: %d/%d", currentLoad, constants.CHANNEL_SIZE))
			
			// 如果当前使用kafka且负载恢复，可以考虑切回channel
			h.modeMutex.RLock()
			usingKafka := h.useKafka
			h.modeMutex.RUnlock()
			
			if usingKafka {
				h.switchToChannel()
			}
		}
	}
}

// switchToKafka 切换到kafka模式
func (h *HybridServer) switchToKafka() {
	h.modeMutex.Lock()
	defer h.modeMutex.Unlock()
	
	if !h.useKafka {
		h.useKafka = true
		zlog.Info("切换到Kafka模式处理消息")
		
		// 将channel中积压的消息转移到kafka
		go h.drainChannelToKafka()
	}
}

// switchToChannel 切换到channel模式
func (h *HybridServer) switchToChannel() {
	h.modeMutex.Lock()
	defer h.modeMutex.Unlock()
	
	if h.useKafka {
		h.useKafka = false
		zlog.Info("切换回Channel模式处理消息")
	}
}

// drainChannelToKafka 将channel中的消息转移到kafka
func (h *HybridServer) drainChannelToKafka() {
	ctx := context.Background()
	drained := 0
	
	for {
		select {
		case data := <-h.Transmit:
			// 发送到kafka
			if err := myKafka.KafkaService.ChatWriter.WriteMessages(ctx, kafka.Message{
				Key:   []byte(strconv.Itoa(config.GetConfig().KafkaConfig.Partition)),
				Value: data,
			}); err != nil {
				zlog.Error(fmt.Sprintf("转移消息到Kafka失败: %v", err))
				// 如果kafka发送失败，重新放回channel
				select {
				case h.Transmit <- data:
				default:
					zlog.Error("Channel已满，消息丢失")
				}
			} else {
				drained++
			}
		default:
			// channel已空
			if drained > 0 {
				zlog.Info(fmt.Sprintf("成功转移%d条消息到Kafka", drained))
			}
			return
		}
	}
}

// SendMessageToTransmit 发送消息到传输通道（支持动态路由）
func (h *HybridServer) SendMessageToTransmit(message []byte) {
	h.modeMutex.RLock()
	usingKafka := h.useKafka
	h.modeMutex.RUnlock()
	
	if usingKafka {
		// 使用kafka发送
		ctx := context.Background()
		if err := myKafka.KafkaService.ChatWriter.WriteMessages(ctx, kafka.Message{
			Key:   []byte(strconv.Itoa(config.GetConfig().KafkaConfig.Partition)),
			Value: message,
		}); err != nil {
			zlog.Error(fmt.Sprintf("Kafka发送失败，回退到Channel: %v", err))
			// kafka发送失败，回退到channel
			select {
			case h.Transmit <- message:
				zlog.Info("消息已回退到Channel")
			default:
				zlog.Error("Channel已满，消息发送失败")
			}
		} else {
			zlog.Debug("消息已通过Kafka发送")
		}
	} else {
		// 使用channel发送
		select {
		case h.Transmit <- message:
			zlog.Debug("消息已通过Channel发送")
		default:
			zlog.Error("Channel已满，消息发送失败")
		}
	}
}

// startKafkaMessageReader 启动kafka消息读取
func (h *HybridServer) startKafkaMessageReader() {
	defer func() {
		if r := recover(); r != nil {
			zlog.Error(fmt.Sprintf("kafka message reader panic: %v", r))
		}
	}()

	for {
		kafkaMessage, err := myKafka.KafkaService.ChatReader.ReadMessage(context.Background())
		if err != nil {
			zlog.Error(err.Error())
			continue
		}
		
		zlog.Info(fmt.Sprintf("从Kafka接收消息: topic=%s, partition=%d, offset=%d", 
			kafkaMessage.Topic, kafkaMessage.Partition, kafkaMessage.Offset))
		
		// 处理kafka消息
		h.processMessage(kafkaMessage.Value)
	}
}

// processMessage 处理消息（统一的消息处理逻辑）
func (h *HybridServer) processMessage(data []byte) {
	var chatMessageReq request.ChatMessageRequest
	if err := json.Unmarshal(data, &chatMessageReq); err != nil {
		zlog.Error(err.Error())
		return
	}

	// 这里复用原有的消息处理逻辑
	if chatMessageReq.Type == message_type_enum.Text {
		h.processTextMessage(chatMessageReq)
	} else if chatMessageReq.Type == message_type_enum.File {
		h.processFileMessage(chatMessageReq)
	} else if chatMessageReq.Type == message_type_enum.AudioOrVideo {
		h.processAVMessage(chatMessageReq)
	}
}

// processTextMessage 处理文本消息
func (h *HybridServer) processTextMessage(chatMessageReq request.ChatMessageRequest) {
	// 存储消息到数据库
	message := model.Message{
		Uuid:       fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11)),
		SessionId:  chatMessageReq.SessionId,
		Type:       chatMessageReq.Type,
		Content:    chatMessageReq.Content,
		Url:        "",
		SendId:     chatMessageReq.SendId,
		SendName:   chatMessageReq.SendName,
		SendAvatar: chatMessageReq.SendAvatar,
		ReceiveId:  chatMessageReq.ReceiveId,
		FileSize:   "0B",
		FileType:   "",
		FileName:   "",
		Status:     message_status_enum.Unsent,
		CreatedAt:  time.Now(),
		AVdata:     "",
	}
	
	// 标准化头像路径
	message.SendAvatar = normalizePath(message.SendAvatar)
	
	if res := dao.GormDB.Create(&message); res.Error != nil {
		zlog.Error(res.Error.Error())
		return
	}

	// 发送消息给接收者
	if message.ReceiveId[0] == 'U' {
		h.sendToUser(message, chatMessageReq)
	} else if message.ReceiveId[0] == 'G' {
		h.sendToGroup(message, chatMessageReq)
	}
}

// processFileMessage 处理文件消息
func (h *HybridServer) processFileMessage(chatMessageReq request.ChatMessageRequest) {
	// 类似processTextMessage的逻辑，但处理文件相关字段
	message := model.Message{
		Uuid:       fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11)),
		SessionId:  chatMessageReq.SessionId,
		Type:       chatMessageReq.Type,
		Content:    "",
		Url:        chatMessageReq.Url,
		SendId:     chatMessageReq.SendId,
		SendName:   chatMessageReq.SendName,
		SendAvatar: chatMessageReq.SendAvatar,
		ReceiveId:  chatMessageReq.ReceiveId,
		FileSize:   chatMessageReq.FileSize,
		FileType:   chatMessageReq.FileType,
		FileName:   chatMessageReq.FileName,
		Status:     message_status_enum.Unsent,
		CreatedAt:  time.Now(),
		AVdata:     "",
	}
	
	message.SendAvatar = normalizePath(message.SendAvatar)
	
	if res := dao.GormDB.Create(&message); res.Error != nil {
		zlog.Error(res.Error.Error())
		return
	}

	if message.ReceiveId[0] == 'U' {
		h.sendToUser(message, chatMessageReq)
	} else if message.ReceiveId[0] == 'G' {
		h.sendToGroup(message, chatMessageReq)
	}
}

// processAVMessage 处理音视频消息
func (h *HybridServer) processAVMessage(chatMessageReq request.ChatMessageRequest) {
	var avData request.AVData
	if err := json.Unmarshal([]byte(chatMessageReq.AVdata), &avData); err != nil {
		zlog.Error(err.Error())
		return
	}

	message := model.Message{
		Uuid:       fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11)),
		SessionId:  chatMessageReq.SessionId,
		Type:       chatMessageReq.Type,
		Content:    "",
		Url:        "",
		SendId:     chatMessageReq.SendId,
		SendName:   chatMessageReq.SendName,
		SendAvatar: chatMessageReq.SendAvatar,
		ReceiveId:  chatMessageReq.ReceiveId,
		FileSize:   "",
		FileType:   "",
		FileName:   "",
		Status:     message_status_enum.Unsent,
		CreatedAt:  time.Now(),
		AVdata:     chatMessageReq.AVdata,
	}

	if avData.MessageId == "PROXY" && (avData.Type == "start_call" || avData.Type == "receive_call" || avData.Type == "reject_call") {
		message.SendAvatar = normalizePath(message.SendAvatar)
		if res := dao.GormDB.Create(&message); res.Error != nil {
			zlog.Error(res.Error.Error())
		}
	}

	if chatMessageReq.ReceiveId[0] == 'U' {
		h.sendAVToUser(message, chatMessageReq)
	}
}

// sendToUser 发送消息给用户
func (h *HybridServer) sendToUser(message model.Message, chatMessageReq request.ChatMessageRequest) {
	messageRsp := respond.GetMessageListRespond{
		SendId:     message.SendId,
		SendName:   message.SendName,
		SendAvatar: chatMessageReq.SendAvatar,
		ReceiveId:  message.ReceiveId,
		Type:       message.Type,
		Content:    message.Content,
		Url:        message.Url,
		FileSize:   message.FileSize,
		FileName:   message.FileName,
		FileType:   message.FileType,
		CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	
	jsonMessage, err := json.Marshal(messageRsp)
	if err != nil {
		zlog.Error(err.Error())
		return
	}
	
	messageBack := &MessageBack{
		Message: jsonMessage,
		Uuid:    message.Uuid,
	}
	
	h.mutex.Lock()
	if receiveClient, ok := h.Clients[message.ReceiveId]; ok {
		receiveClient.SendBack <- messageBack
	}
	if sendClient, ok := h.Clients[message.SendId]; ok {
		sendClient.SendBack <- messageBack
	}
	h.mutex.Unlock()

	// 更新Redis缓存
	h.updateRedisCache("message_list_"+message.SendId+"_"+message.ReceiveId, messageRsp)
}

// sendToGroup 发送消息给群组
func (h *HybridServer) sendToGroup(message model.Message, chatMessageReq request.ChatMessageRequest) {
	messageRsp := respond.GetGroupMessageListRespond{
		SendId:     message.SendId,
		SendName:   message.SendName,
		SendAvatar: chatMessageReq.SendAvatar,
		ReceiveId:  message.ReceiveId,
		Type:       message.Type,
		Content:    message.Content,
		Url:        message.Url,
		FileSize:   message.FileSize,
		FileName:   message.FileName,
		FileType:   message.FileType,
		CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	
	jsonMessage, err := json.Marshal(messageRsp)
	if err != nil {
		zlog.Error(err.Error())
		return
	}
	
	messageBack := &MessageBack{
		Message: jsonMessage,
		Uuid:    message.Uuid,
	}
	
	// 获取群组成员
	var group model.GroupInfo
	if res := dao.GormDB.Where("uuid = ?", message.ReceiveId).First(&group); res.Error != nil {
		zlog.Error(res.Error.Error())
		return
	}
	
	var members []string
	if err := json.Unmarshal(group.Members, &members); err != nil {
		zlog.Error(err.Error())
		return
	}
	
	h.mutex.Lock()
	for _, member := range members {
		if client, ok := h.Clients[member]; ok {
			client.SendBack <- messageBack
		}
	}
	h.mutex.Unlock()

	// 更新Redis缓存
	h.updateRedisCache("group_messagelist_"+message.ReceiveId, messageRsp)
}

// sendAVToUser 发送音视频消息给用户
func (h *HybridServer) sendAVToUser(message model.Message, chatMessageReq request.ChatMessageRequest) {
	messageRsp := respond.AVMessageRespond{
		SendId:     message.SendId,
		SendName:   message.SendName,
		SendAvatar: message.SendAvatar,
		ReceiveId:  message.ReceiveId,
		Type:       message.Type,
		Content:    message.Content,
		Url:        message.Url,
		FileSize:   message.FileSize,
		FileName:   message.FileName,
		FileType:   message.FileType,
		CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
		AVdata:     message.AVdata,
	}
	
	jsonMessage, err := json.Marshal(messageRsp)
	if err != nil {
		zlog.Error(err.Error())
		return
	}
	
	messageBack := &MessageBack{
		Message: jsonMessage,
		Uuid:    message.Uuid,
	}
	
	h.mutex.Lock()
	if receiveClient, ok := h.Clients[message.ReceiveId]; ok {
		receiveClient.SendBack <- messageBack
	}
	h.mutex.Unlock()
}

// updateRedisCache 更新Redis缓存
func (h *HybridServer) updateRedisCache(key string, messageRsp interface{}) {
	rspString, err := myredis.GetKeyNilIsErr(key)
	if err == nil {
		// 缓存存在，更新
		var rsp []interface{}
		if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
			zlog.Error(err.Error())
			return
		}
		rsp = append(rsp, messageRsp)
		rspByte, err := json.Marshal(rsp)
		if err != nil {
			zlog.Error(err.Error())
			return
		}
		if err := myredis.SetKeyEx(key, string(rspByte), time.Minute*constants.REDIS_TIMEOUT); err != nil {
			zlog.Error(err.Error())
		}
	} else {
		if !errors.Is(err, redis.Nil) {
			zlog.Error(err.Error())
		}
	}
}

// GetCurrentMode 获取当前消息处理模式
func (h *HybridServer) GetCurrentMode() string {
	h.modeMutex.RLock()
	defer h.modeMutex.RUnlock()
	
	if h.useKafka {
		return "kafka"
	}
	return "channel"
}

// GetChannelStatus 获取channel状态信息
func (h *HybridServer) GetChannelStatus() map[string]interface{} {
	h.channelMonitor.mutex.RLock()
	defer h.channelMonitor.mutex.RUnlock()
	
	status := map[string]interface{}{
		"current_load":      len(h.Transmit),
		"max_capacity":      constants.CHANNEL_SIZE,
		"load_percentage":   float64(len(h.Transmit)) / float64(constants.CHANNEL_SIZE) * 100,
		"is_overloaded":     h.channelMonitor.isOverloaded,
		"threshold_ratio":   h.channelMonitor.thresholdRatio,
		"current_mode":      h.GetCurrentMode(),
	}
	
	if h.channelMonitor.overloadStartTime != nil {
		status["overload_duration"] = time.Since(*h.channelMonitor.overloadStartTime).Seconds()
	}
	
	return status
}

// Close 关闭混合服务器
func (h *HybridServer) Close() {
	zlog.Info("正在关闭混合服务器...")
}

// SendClientToLogin 发送客户端到登录通道
func (h *HybridServer) SendClientToLogin(client *Client) {
	h.Login <- client
}

// SendClientToLogout 发送客户端到登出通道
func (h *HybridServer) SendClientToLogout(client *Client) {
	h.Logout <- client
}