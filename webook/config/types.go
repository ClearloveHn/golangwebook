package config

// 配置信息

// 顶层的配置结构体,用于表示整个应用的配置信息
type config struct {
	DB    DBConfig
	Redis RedisConfig
}

// DBConfig 表示数据库的配置信息。
type DBConfig struct {
	DSN string // 数据库的连接字符串
}

// RedisConfig Redis 的配置信息
type RedisConfig struct {
	Addr string // Redis 服务器的地址
}
