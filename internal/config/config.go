package config

import (
	"github.com/BurntSushi/toml"
	"log"
	"time"
)

type MainConfig struct {
	AppName string `toml:"appName"`
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
}

type MysqlConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
	DatabaseName string `toml:"databaseName"`
}

type RedisConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	Db       int    `toml:"db"`
}

type AuthCodeConfig struct {
	AccessKeyID     string `toml:"accessKeyID"`
	AccessKeySecret string `toml:"accessKeySecret"`
	SignName        string `toml:"signName"`
	TemplateCode    string `toml:"templateCode"`
}

// 号码认证服务配置（推荐使用）
type SmsAuthConfig struct {
	AccessKeyID     string `toml:"accessKeyID"`
	AccessKeySecret string `toml:"accessKeySecret"`
	UseNewService   bool   `toml:"useNewService"` // 是否使用新的号码认证服务
}

type LogConfig struct {
	LogPath string `toml:"logPath"`
}

type KafkaConfig struct {
	MessageMode           string        `toml:"messageMode"`
	HostPort              string        `toml:"hostPort"`
	LoginTopic            string        `toml:"loginTopic"`
	LogoutTopic           string        `toml:"logoutTopic"`
	ChatTopic             string        `toml:"chatTopic"`
	Partition             int           `toml:"partition"`
	Timeout               time.Duration `toml:"timeout"`
	// 混合模式配置参数
	HybridThreshold       float64       `toml:"hybridThreshold"`       // 触发切换的阈值比例 (0.0-1.0)
	HybridMonitorInterval int           `toml:"hybridMonitorInterval"` // 监控间隔时间(秒)
	HybridSwitchDuration  int           `toml:"hybridSwitchDuration"`  // 持续时间阈值(秒)
}

type StaticSrcConfig struct {
	StaticAvatarPath string `toml:"staticAvatarPath"`
	StaticFilePath   string `toml:"staticFilePath"`
}

type Config struct {
	MainConfig      `toml:"mainConfig"`
	MysqlConfig     `toml:"mysqlConfig"`
	RedisConfig     `toml:"redisConfig"`
	AuthCodeConfig  `toml:"authCodeConfig"`
	SmsAuthConfig   `toml:"smsAuthConfig"`   // 新增号码认证服务配置
	LogConfig       `toml:"logConfig"`
	KafkaConfig     `toml:"kafkaConfig"`
	StaticSrcConfig `toml:"staticSrcConfig"`
}

var config *Config

func LoadConfig() error {
	// 本地部署 - Windows
	if _, err := toml.DecodeFile("configs/config.toml", config); err != nil {
		log.Fatal(err.Error())
		return err
	}
	// Ubuntu22.04云服务器部署
	// if _, err := toml.DecodeFile("/root/project/KamaChat/configs/config_local.toml", config); err != nil {
	// 	log.Fatal(err.Error())
	// 	return err
	// }
	return nil
}

func GetConfig() *Config {
	if config == nil {
		config = new(Config)
		_ = LoadConfig()
	}
	return config
}
