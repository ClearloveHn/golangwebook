//go:build !k8s

package config

// Config config结构体的初始化和实例化
var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(localhost:13316)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
