package dao

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// MongoDBArticleDAO 是 ArticleDAO 接口的 MongoDB 实现
type MongoDBArticleDAO struct {
	node    *snowflake.Node   // 雪花算法节点,用于生成唯一ID
	col     *mongo.Collection // 文章集合
	liveCol *mongo.Collection // 已发布文章集合
}

var _ ArticleDAO = &MongoDBArticleDAO{}

func NewMongoDBArticleDAO(mdb *mongo.Database, node *snowflake.Node) *MongoDBArticleDAO {
	return &MongoDBArticleDAO{
		node:    node,
		liveCol: mdb.Collection("published_articles"),
		col:     mdb.Collection("articles"),
	}
}

// Insert 插入一篇新文章
func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	art.Id = m.node.Generate().Int64() // 生成唯一ID

	_, err := m.col.InsertOne(ctx, art) // 插入文章到 MongoDB

	return art.Id, err
}

// UpdateById 根据ID更新文章
func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()

	filter := bson.D{bson.E{Key: "id", Value: art.Id}, // 构建过滤器,匹配文章ID
		bson.E{Key: "author_id", Value: art.AuthorId}} // 和作者ID

	set := bson.D{bson.E{Key: "$set", Value: bson.M{ // 构建更新文档
		"title":   art.Title,   // 更新标题
		"content": art.Content, // 更新内容
		"status":  art.Status,  // 更新状态
		"utime":   now,         // 更新更新时间
	}}}

	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.New("article not found")
	}

	return nil
}

// Sync 同步文章
func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)

	if id > 0 { // 如果文章ID大于0,说明是更新操作
		err = m.UpdateById(ctx, art) // 更新文章
	} else { // 否则,是插入操作
		id, err = m.Insert(ctx, art) // 插入文章,并获取新文章的ID
	}
	if err != nil {
		return 0, err // 如果出错,返回错误
	}

	art.Id = id
	now := time.Now().UnixMilli()
	art.Utime = now

	filter := bson.D{bson.E{Key: "id", Value: art.Id}, // 构建过滤器,匹配文章ID
		bson.E{Key: "author_id", Value: art.AuthorId}} // 和作者ID

	set := bson.D{bson.E{Key: "$set", Value: art}, // 构建更新文档,设置整个文章
		bson.E{Key: "$setOnInsert", // 如果是插入操作,设置创建时间
			Value: bson.D{bson.E{Key: "ctime", Value: now}}}}

	_, err = m.liveCol.UpdateOne(ctx, // 更新或插入文章到已发布文章集合
		filter, set, // 过滤器和更新文档
		options.Update().SetUpsert(true)) // 设置 upsert 为 true,表示如果不存在则插入

	return id, err
}

// SyncStatus 同步文章状态
func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {

	filter := bson.D{bson.E{Key: "id", Value: id}, // 构建过滤器,匹配文章ID
		bson.E{Key: "author_id", Value: uid}} // 和作者ID

	sets := bson.D{bson.E{Key: "$set", // 构建更新文档
		Value: bson.D{bson.E{Key: "status", Value: status}}}} // 设置状态

	res, err := m.col.UpdateOne(ctx, filter, sets) // 更新文章状态

	if err != nil {
		return err
	}

	if res.ModifiedCount != 1 {
		return errors.New("ID 不对或者创作者不对")
	}

	_, err = m.liveCol.UpdateOne(ctx, filter, sets) // 更新已发布文章的状态

	return err
}

func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}
