package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:generate mockgen -source=./article.go -package=cachemocks -destination=./mocks/article.mock.go ArticleCache

// ArticleCache 是一个接口，定义了文章缓存的相关操作
type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, res []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, res domain.Article) error
}

// ArticleRedisCache 是一个结构体，实现了 ArticleCache 接口，使用 Redis 作为缓存存储
type ArticleRedisCache struct {
	client redis.Cmdable
}

// NewArticleRedisCache 是一个函数，用于创建 ArticleRedisCache 实例
func NewArticleRedisCache(client redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{
		client: client,
	}
}

// GetFirstPage 方法用于获取某个用户的首页文章缓存
func (a *ArticleRedisCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.firstKey(uid)

	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

// SetFirstPage 方法用于设置某个用户的首页文章缓存
func (a *ArticleRedisCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}

	key := a.firstKey(uid)
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}

	return a.client.Set(ctx, key, string(val), 0).Err()
}

// DelFirstPage 方法用于删除某个用户的首页文章缓存
func (a *ArticleRedisCache) DelFirstPage(ctx context.Context, uid int64) error {
	return a.client.Del(ctx, a.firstKey(uid)).Err()
}

// Get 方法用于获取某篇文章的缓存
func (a *ArticleRedisCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}

	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

// Set 方法用于设置某篇文章的缓存
func (a *ArticleRedisCache) Set(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}

	return a.client.Set(ctx, a.key(art.Id), val, time.Minute*10).Err()
}

// GetPub 方法用于获取某篇已发布文章的缓存
func (a *ArticleRedisCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}

	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

// SetPub 方法用于设置某篇已发布文章的缓存
func (a *ArticleRedisCache) SetPub(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}

	return a.client.Set(ctx, a.key(art.Id), val, 0).Err()
}

// pubKey 方法用于生成已发布文章的缓存键
func (a *ArticleRedisCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}

// key 方法用于生成文章的缓存键
func (a *ArticleRedisCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}

// firstKey 方法用于生成用户首页文章的缓存键
func (a *ArticleRedisCache) firstKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}
