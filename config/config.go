package config

import (
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Global 全局配置实例
var Global AppConfig

// MySQL 数据库配置
type MySQL struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	MaxCon   int    `mapstructure:"max_con"` // 连接池最大连接数
}

// Redis 配置
type Redis struct {
	Addr        string `mapstructure:"addr"`
	Password    string `mapstructure:"password"`
	DB          int    `mapstructure:"db"`
	PoolSize    int    `mapstructure:"poolSize"`
	MinIdleConn int    `mapstructure:"minIdleConn"`
}

// OSS 阿里云对象存储配置
type OSS struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accessKeyId"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
	Bucket          string `mapstructure:"bucket"`
}

// Timeout 心跳/超时配置
type Timeout struct {
	DelayHeartbeat   int `mapstructure:"delayHeartbeat"`
	HeartbeatHz      int `mapstructure:"heartbeatHz"`
	HeartbeatMaxTime int `mapstructure:"heartbeatMaxTime"`
	RedisOnlineTime  int `mapstructure:"redisOnlineTime"`
}

// Port 端口配置
type Port struct {
	Server string `mapstructure:"server"`
}

// Log 日志配置
type Log struct {
	Output     string `mapstructure:"output"`
	File       string `mapstructure:"file"`
	FormatJson bool   `mapstructure:"format_json"`
}

// Jwt JWT配置
type Jwt struct {
	AccessTokenSecret string `mapstructure:"accessSecret"`
	AccessExpireTime  int    `mapstructure:"accessExpire"`
}

// AppConfig 总配置
type AppConfig struct {
	MySQL   MySQL         `mapstructure:"mysql"`
	Redis   Redis         `mapstructure:"redis"`
	OSS     OSS           `mapstructure:"oss"`
	Timeout Timeout       `mapstructure:"timeout"`
	Port    Port          `mapstructure:"port"`
	Log     Log           `mapstructure:"log"`
	DB      *gorm.DB      `mapstructure:"-"`
	Red     *redis.Client `mapstructure:"-"`
	Jwt     Jwt           `mapstructure:"jwt"`
}

func LoadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Faild to Read config:", err)
	}
	if err := viper.Unmarshal(&Global); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

}
