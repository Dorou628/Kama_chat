package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	myconfig "kama_chat_server/internal/config"
	"kama_chat_server/pkg/zlog"
	"time"
)

var ctx = context.Background()

type kafkaService struct {
	ChatWriter      *kafka.Writer
	ChatReader      *kafka.Reader
	AsyncTaskWriter *kafka.Writer
	AsyncTaskReader *kafka.Reader
	KafkaConn       *kafka.Conn
}

var KafkaService = new(kafkaService)

// KafkaInit 初始化kafka
func (k *kafkaService) KafkaInit() {
	//k.CreateTopic()
	kafkaConfig := myconfig.GetConfig().KafkaConfig
	k.ChatWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaConfig.HostPort),
		Topic:                  kafkaConfig.ChatTopic,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           kafkaConfig.Timeout * time.Second,
		RequiredAcks:           kafka.RequireNone,
		AllowAutoTopicCreation: false,
	}
	k.ChatReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaConfig.HostPort},
		Topic:          kafkaConfig.ChatTopic,
		CommitInterval: kafkaConfig.Timeout * time.Second,
		GroupID:        "chat",
		StartOffset:    kafka.LastOffset,
	})
	
	// 初始化异步任务相关的Writer和Reader
	k.AsyncTaskWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaConfig.HostPort),
		Topic:                  "async_tasks",
		Balancer:               &kafka.Hash{},
		WriteTimeout:           kafkaConfig.Timeout * time.Second,
		RequiredAcks:           kafka.RequireNone,
		AllowAutoTopicCreation: false,
	}
	k.AsyncTaskReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaConfig.HostPort},
		Topic:          "async_tasks",
		CommitInterval: kafkaConfig.Timeout * time.Second,
		GroupID:        "async_task_workers",
		StartOffset:    kafka.LastOffset,
	})
}

func (k *kafkaService) KafkaClose() {
	if err := k.ChatWriter.Close(); err != nil {
		zlog.Error(err.Error())
	}
	if err := k.ChatReader.Close(); err != nil {
		zlog.Error(err.Error())
	}
	if err := k.AsyncTaskWriter.Close(); err != nil {
		zlog.Error(err.Error())
	}
	if err := k.AsyncTaskReader.Close(); err != nil {
		zlog.Error(err.Error())
	}
}

// CreateTopic 创建topic
func (k *kafkaService) CreateTopic() {
	// 如果已经有topic了，就不创建了
	kafkaConfig := myconfig.GetConfig().KafkaConfig

	chatTopic := kafkaConfig.ChatTopic

	// 连接至任意kafka节点
	var err error
	k.KafkaConn, err = kafka.Dial("tcp", kafkaConfig.HostPort)
	if err != nil {
		zlog.Error(err.Error())
	}

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             chatTopic,
			NumPartitions:     kafkaConfig.Partition,
			ReplicationFactor: 1,
		},
		{
			Topic:             "async_tasks",
			NumPartitions:     kafkaConfig.Partition,
			ReplicationFactor: 1,
		},
	}

	// 创建topic
	if err = k.KafkaConn.CreateTopics(topicConfigs...); err != nil {
		zlog.Error(err.Error())
	}

}
