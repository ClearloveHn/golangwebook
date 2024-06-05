package service

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/ClearloveHn/golangwebook/webook/internal/events/article"
	"github.com/ClearloveHn/golangwebook/webook/internal/repository"
	"gorm.io/gorm/logger"
	"time"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=./mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer
	userRepo repository.UserRepository
}

func NewArticleService(repo repository.ArticleRepository,
	producer article.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

// Save 方法保存文章
func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished // 设置文章状态为未发布
	if art.Id > 0 {
		err := a.repo.Update(ctx, art) // 更新已有文章
		return art.Id, err
	}
	return a.repo.Create(ctx, art) // 创建新文章
}

// Publish 方法发布文章
func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished // 设置文章状态为已发布
	return a.repo.Sync(ctx, art)               // 同步文章到仓库
}

// Withdraw 方法撤回文章
func (a *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate) // 同步文章状态为私有
}

// GetByAuthor 方法根据作者获取文章
func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

// GetById 方法根据文章 ID 获取文章
func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

// GetPubById 方法根据文章 ID 获取已发布的文章，并异步发送阅读事件
func (a *articleService) GetPubById(ctx context.Context, id, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)
	go func() {
		if err == nil {
			// 在这里发送一个阅读事件
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(err))
			}
		}
	}()

	return res, err
}

// ListPub 方法获取指定时间之后的已发布文章列表
func (a *articleService) ListPub(ctx context.Context,
	start time.Time, offset, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, start, offset, limit)
}
