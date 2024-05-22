//go:build k8s

package config

// Config config结构体的初始化和实例化
var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(webook-record-mysql:3308)/webook",
	},
	Redis: RedisConfig{
		Addr: "webook-record-redis:6379",
	},
}
