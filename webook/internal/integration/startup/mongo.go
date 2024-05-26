package startup

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// MongoDB连接的初始化功能

// InitMongoDB 是一个函数,用于初始化 MongoDB 连接
func InitMongoDB() *mongo.Database {
	// 创建一个带超时的 context,用于控制连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建一个 MongoDB 命令监视器,用于监视 MongoDB 命令的执行情况
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			// 在命令开始执行时打印命令信息
			fmt.Println(evt.Command)
		},
	}

	// 创建 MongoDB 客户端选项,设置连接 URI 和命令监视器
	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(monitor)

	// 使用客户端选项连接到 MongoDB
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		// 如果连接出错,则抛出 panic
		panic(err)
	}

	// 返回连接的 MongoDB 数据库对象,数据库名称为 "webook"
	return client.Database("webook")
}
