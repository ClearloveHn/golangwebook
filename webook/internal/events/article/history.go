package article

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/IBM/sarama"
	"time"
)

// 记录历史事件

type HistoryRecordConsumer struct {
	repo   repository.HistoryRecordRepository
	client sarama.Client // Sarama客户端,用于连接和消费Kafka消息
	l      logger.LoggerV1
}

// Start 用于启动消费者,开始消费和处理历史记录事件
func (i *HistoryRecordConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}

	go func() {
		// 使用 samarax 包的 NewHandler 函数创建一个单个消息处理器,用于逐个处理消息
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			samarax.NewHandler[ReadEvent](i.l, i.Consume))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()

	return err
}

func (i *HistoryRecordConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 添加一条历史记录
	return i.repo.AddRecord(ctx, domain.HistoryRecord{
		BizId: event.Aid,
		Biz:   "article",
		Uid:   event.Uid,
	})

}
