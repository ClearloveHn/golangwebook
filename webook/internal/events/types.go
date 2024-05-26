package events

// 启动消费者模型接口

// Consumer 是一个接口类型,表示事件消费者
// 实现了 Consumer 接口的类型可以作为事件消费者,用于消费和处理事件
type Consumer interface {
	// Start 是 Consumer 接口中定义的一个方法
	//该方法用于启动事件消费者,开始消费和处理事件
	Start() error
}
