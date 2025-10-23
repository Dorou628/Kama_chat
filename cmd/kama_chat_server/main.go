package main

import (
	"fmt"
	"kama_chat_server/internal/config"
	"kama_chat_server/internal/https_server"
	"kama_chat_server/internal/service/chat"
	"kama_chat_server/internal/service/kafka"
	myredis "kama_chat_server/internal/service/redis"
	"kama_chat_server/pkg/zlog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conf := config.GetConfig()
	host := conf.MainConfig.Host
	port := conf.MainConfig.Port
	kafkaConfig := conf.KafkaConfig
	
	// 根据消息模式初始化相应的服务
	if kafkaConfig.MessageMode == "kafka" || kafkaConfig.MessageMode == "hybrid" {
		kafka.KafkaService.KafkaInit()
	}

	if kafkaConfig.MessageMode == "channel" {
		go chat.ChatServer.Start()
	} else if kafkaConfig.MessageMode == "hybrid" {
		go chat.HybridChatServer.Start()
	} else {
		go chat.KafkaChatServer.Start()
	}

	go func() {
		// 本地开发模式 - 使用HTTP
		// if err := https_server.GE.Run(fmt.Sprintf("%s:%d", host, port)); err != nil {
		// 	zlog.Fatal("server running fault")
		// 	return
		// }
		// Win10本地部署 - HTTPS模式（需要SSL证书时启用）
		if err := https_server.GE.RunTLS(fmt.Sprintf("%s:%d", host, port), "pkg/ssl/127.0.0.1+2.pem", "pkg/ssl/127.0.0.1+2-key.pem"); err != nil {
			zlog.Fatal("server running fault")
			return
		}
		// Ubuntu22.04云服务器部署
		// if err := https_server.GE.RunTLS(fmt.Sprintf("%s:%d", host, port), "/etc/ssl/certs/server.crt", "/etc/ssl/private/server.key"); err != nil {
		// 	zlog.Fatal("server running fault")
		// 	return
		// }
	}()

	// 设置信号监听
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-quit

	// 根据模式关闭相应的服务
	if kafkaConfig.MessageMode == "kafka" || kafkaConfig.MessageMode == "hybrid" {
		kafka.KafkaService.KafkaClose()
	}

	if kafkaConfig.MessageMode == "channel" {
		chat.ChatServer.Close()
	} else if kafkaConfig.MessageMode == "hybrid" {
		chat.HybridChatServer.Close()
	}

	zlog.Info("关闭服务器...")

	// 删除所有Redis键
	if err := myredis.DeleteAllRedisKeys(); err != nil {
		zlog.Error(err.Error())
	} else {
		zlog.Info("所有Redis键已删除")
	}

	zlog.Info("服务器已关闭")
}
