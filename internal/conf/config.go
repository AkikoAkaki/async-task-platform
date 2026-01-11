package conf

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig    `mapstructure:"app"`
	Server ServerConfig `mapstructure:"server"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Queue  QueueConfig  `mapstructure:"queue"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type ServerConfig struct {
	Port     int `mapstructure:"port"`
	GrpcPort int `mapstructure:"grpc_port"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type QueueConfig struct {
	// 任务在 Running 状态的超时时间 (秒)，超过此时间未 ACK 则被 Watchdog 恢复
	VisibilityTimeout int `mapstructure:"visibility_timeout"`
	// Watchdog 扫描间隔 (秒)
	WatchdogInterval int `mapstructure:"watchdog_interval"`
	// 默认最大重试次数
	MaxRetries int `mapstructure:"max_retries"`
}

// Load 加载配置。
// 优先级：环境变量 > 配置文件 > 默认值
func Load(path string) (*Config, error) {
	v := viper.New()

	// 1. 设置配置文件路径
	v.AddConfigPath(path)     // 比如 "./config"
	v.SetConfigName("config") // 文件名 config (不带后缀)
	v.SetConfigType("yaml")

	// 2. 开启环境变量自动匹配
	// 比如 config.yaml 里 redis.addr 对应环境变量 DDQ_REDIS_ADDR
	v.SetEnvPrefix("DDQ")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 3. 读取配置
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using environment variables only")
		} else {
			return nil, err
		}
	}

	// 4. 绑定到结构体
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
