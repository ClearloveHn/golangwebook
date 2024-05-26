package startup

import "github.com/IBM/sarama"

// 初始化Sarama客户端和同步生产者

// InitSaramaClient 是一个函数,用于初始化 Sarama 客户端
func InitSaramaClient() sarama.Client {
	// 创建一个新的 Sarama 配置对象
	scfg := sarama.NewConfig()

	// 设置生产者成功发送消息后返回成功响应
	scfg.Producer.Return.Successes = true

	// 使用指定的 Kafka 服务器地址和配置创建 Sarama 客户端
	clint, err := sarama.NewClient([]string{"localhost:9094"}, scfg)
	if err != nil {
		// 如果创建客户端出错,则抛出 panic
		panic(err)
	}

	// 返回初始化后的 Sarama 客户端对象
	return clint
}

// InitSyncProducer 是一个函数,用于初始化 Sarama 同步生产者
func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	// 使用指定的 Sarama 客户端创建同步生产者
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		// 如果创建同步生产者出错,则抛出 panic
		panic(err)
	}

	// 返回初始化后的 Sarama 同步生产者对象
	return p
}
