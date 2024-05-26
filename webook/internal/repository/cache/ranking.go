package cache

import (
	"context"
	"encoding/json"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

// RankingCache 接口,包含排行榜缓存的相关操作
type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingRedisCache struct {
	client     redis.Cmdable // 类型为 redis.Cmdable,用于执行 Redis 命令
	key        string        // 表示排行榜缓存的键名
	expiration time.Duration // 表示排行榜缓存的过期时间
}

func NewRankingRedisCache(client redis.Cmdable) RankingCache {
	return &RankingRedisCache{
		client:     client,
		key:        "ranking:top_n",
		expiration: time.Minute * 3,
	}
}

// Set 方法,用于设置排行榜缓存
func (r *RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := range arts { // 遍历文章切片
		arts[i].Content = arts[i].Abstract() // 将文章内容替换为文章摘要
	}

	val, err := json.Marshal(arts) // 将文章切片序列化为 JSON 格式
	if err != nil {                // 如果序列化出错
		return err // 返回错误
	}

	return r.client.Set(ctx, r.key, string(val), r.expiration).Err() // 调用 Redis 的 Set 方法设置排行榜缓存,并设置过期时间
}

// Get 方法,用于获取排行榜缓存
func (r *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := r.client.Get(ctx, r.key).Bytes() // 调用 Redis 的 Get 方法获取排行榜缓存的值
	if err != nil {
		return nil, err
	}

	var res []domain.Article
	err = json.Unmarshal(val, &res)

	return res, err
}
