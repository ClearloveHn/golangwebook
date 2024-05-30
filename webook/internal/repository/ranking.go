package repository

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/ClearloveHn/golangwebook/webook/internal/repository/cache"
)

// RankingRepository 定义 RankingRepository 接口
type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error // 替换前 N 名文章
	GetTopN(ctx context.Context) ([]domain.Article, error)        // 获取前 N 名文章
}

type CachedRankingRepository struct {
	cache cache.RankingCache
}

func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

// ReplaceTopN 实现 RankingRepository 接口的 ReplaceTopN 方法
func (repo *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return repo.cache.Set(ctx, arts) // 调用 RankingCache 的 Set 方法替换前 N 名文章
}

// GetTopN 实现 RankingRepository 接口的 GetTopN 方法
func (repo *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return repo.cache.Get(ctx) // 调用 RankingCache 的 Get 方法获取前 N 名文章
}
