package dao

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"time"
)

// InitTables 初始化MySQL数据库表结构
func InitTables(db *gorm.DB) error {
	// 使用AutoMigrate自动迁移数据库表结构,创建或更新表
	// 注意:在生产环境中,自动迁移可能不是最佳实践,应该手动管理数据库表结构
	return db.AutoMigrate(&User{},
		&Article{},
		&PublishedArticle{},
		&Interactive{},
		&UserLikeBiz{},
		&UserCollectionBiz{},
		&AsyncSms{},
		&Job{},
	)
}

// InitCollection 初始化MongoDB集合和索引
func InitCollection(mdb *mongo.Database) error {

	// 创建一个带超时的上下文,确保操作不会无限期阻塞
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 获取 "articles" 集合
	col := mdb.Collection("articles")

	// 为 "articles" 集合创建多个索引
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// 创建 "id" 字段的唯一索引
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},

		{
			// 创建 "author_id" 字段的普通索引
			Keys: bson.D{bson.E{Key: "author_id", Value: 1}},
		},
	})
	if err != nil {
		return err
	}

	// 获取 "published_articles" 集合
	liveCol := mdb.Collection("published_articles")

	// 为 "published_articles" 集合创建多个索引
	_, err = liveCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// 创建 "id" 字段的唯一索引
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			// 创建 "author_id" 字段的普通索引
			Keys: bson.D{bson.E{Key: "author_id", Value: 1}},
		},
	})
	return err

}
