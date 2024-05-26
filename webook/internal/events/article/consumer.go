package article

import (
	"context"
	"github.com/IBM/sarama"
	"time"
)

// 使用 Sarama 客户端和 samarax 包来实现 Kafka 消息的消费和处理

type InteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client // Sarama客户端,用于连接和消费Kafka消息
	l      logger.LoggerV1
}

func NewInteractiveReadEventConsumer(repo repository.InteractiveRepository,
	client sarama.Client, l logger.LoggerV1) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{repo: repo, client: client, l: l}
}

// Start 用于启动消费者
func (consumer *InteractiveReadEventConsumer) Start() error {
	// 使用Sarama客户端创建一个新的消费者组
	cg, err := sarama.NewConsumerGroupFromClient("interactive", consumer.client)
	if err != nil {
		return err
	}

	// 在新的协程中启动消费者组,并消费指定主题的消息
	go func() {
		// 使用 samarax 包的 NewBatchHandler 函数创建一个批量处理器,用于批量处理消息
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			samarax.NewBatchHandler[ReadEvent](i.l, i.BatchConsume))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

// Consume 用于单个消费和处理交互式阅读事件
func (consumer *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 增加阅读计数
	return consumer.repo.IncrReadCnt(ctx, "article", event.Aid)
}

// BatchConsume 用于批量消费和处理交互式阅读事件
func (consumer *InteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage,
	events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, evt.Aid)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return consumer.repo.BatchIncrReadCnt(ctx, bizs, bizIds)
}
