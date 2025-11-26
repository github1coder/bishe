package setting

import (
	"gopkg.in/ini.v1"
)

var Conf = new(AppConfig)

// AppConfig 应用程序配置
type AppConfig struct {
	Release bool        `ini:"release"`
	Port    int         `ini:"port"`
	Redis   RedisConfig `ini:"redis"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string `ini:"host"`
	Port         int    `ini:"port"`
	Password     string `ini:"password"`
	DB           int    `ini:"db"`
	PoolSize     int    `ini:"pool_size"`
	DialTimeout  int    `ini:"dial_timeout"`
	ReadTimeout  int    `ini:"read_timeout"`
	WriteTimeout int    `ini:"write_timeout"`
}

func Init(file string) error {
	return ini.MapTo(Conf, file)
}
