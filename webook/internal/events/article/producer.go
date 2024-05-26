package article

import (
	"encoding/json"
	"github.com/IBM/sarama"
)

// TopicReadEvent 是一个常量,表示文章阅读事件的主题名称
const TopicReadEvent = "article_read"

// Producer 是一个接口,定义了生产阅读事件的方法
type Producer interface {
	ProduceReadEvent(evt ReadEvent) error
}

// ReadEvent 单个文章阅读事件
type ReadEvent struct {
	Aid int64
	Uid int64
}

// BatchReadEvent 批量文章阅读事件
type BatchReadEvent struct {
	Aids []int64
	Uids []int64
}

// SaramaSyncProducer 使用Sarama同步生产者的阅读事件生产者
type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{producer: producer}
}

// ProduceReadEvent 是SaramaSyncProducer的一个方法,用于生产单个阅读事件
func (s *SaramaSyncProducer) ProduceReadEvent(evt ReadEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	// 使用Sarama同步生产者发送阅读事件消息
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicReadEvent,            // 指定主题为文章阅读事件的主题
		Value: sarama.StringEncoder(val), // 将JSON格式的事件数据作为消息的值
	})

	return err
}
