package repository

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/ClearloveHn/golangwebook/webook/internal/repository/cache"
	"github.com/ClearloveHn/golangwebook/webook/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

//go:generate mockgen -source=./article.go -package=repomocks -destination=./mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao   dao.ArticleDAO
	cache cache.ArticleCache

	// 因为如果你直接访问 UserDAO，你就绕开了 repository，
	// repository 一般都有一些缓存机制
	userRepo UserRepository

	readerDAO dao.ArticleReaderDAO
	authorDAO dao.ArticleAuthorDAO
	db        *gorm.DB
}

func NewCachedArticleRepository(dao dao.ArticleDAO, userRepo UserRepository,
	cache cache.ArticleCache) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
	}
}

// Create 创建文章
func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {

	id, err := c.dao.Insert(ctx, c.toEntity(art))

	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id) // 删除第一页缓存
		if er != nil {
			// 也要记录日志
		}
	}

	return id, err
}

// Update 更新文章
func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {

	err := c.dao.UpdateById(ctx, c.toEntity(art))

	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}

	return err
}

// Sync 同步文章
func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {

	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}

	// 在这里尝试，设置缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// 你可以灵活设置过期时间
		user, er := c.userRepo.FindById(ctx, art.Author.Id)
		if er != nil {
			// 要记录日志
			return
		}

		// 设置文章的作者
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}

		// 设置已发布文章的缓存
		er = c.cache.SetPub(ctx, art)
		if er != nil {
			// 记录日志
		}
	}()

	return id, err
}

// SyncStatus 同步文章状态
func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {

	err := c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	if err == nil {
		er := c.cache.DelFirstPage(ctx, uid)
		if er != nil {
			// 也要记录日志
		}
	}

	return err
}

// GetByAuthor 根据作者获取文章
func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 首先第一步，判定要不要查询缓存
	// 事实上， limit <= 100 都可以查询缓存
	if offset == 0 && limit == 100 {
		//if offset == 0 && limit <= 100 {
		res, err := c.cache.GetFirstPage(ctx, uid) // 从缓存中获取作者第一页的文章
		if err == nil {
			return res, err
		} else {
			// 要考虑记录日志
			// 缓存未命中，你是可以忽略的
		}
	}

	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}

	// 切片之间的转换。
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})

	// 写缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			// 缓存回写失败，不一定是大问题，但有可能是大问题
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 记录日志
				// 我需要监控这里
			}
		}
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, res) // 预缓存文章
	}()

	return res, nil
}

// GetById 根据ID获取文章
func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}

	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}

	res = c.toDomain(art)
	go func() {
		er := c.cache.Set(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

// GetPubById 根据ID获取已发布的文章
func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {

	// 从缓存中获取已发布的文章
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}

	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}

	// 我现在要去查询对应的 User 信息，拿到创作者信息
	res = c.toDomain(dao.Article(art))
	author, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
		// 要额外记录日志，因为你吞掉了错误信息
		//return res, nil
	}

	res.Author.Name = author.Nickname
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()

	return res, nil
}

// ListPub 获取已发布的文章列表
func (c *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {

	arts, err := c.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map[dao.PublishedArticle, domain.Article](arts,
		func(idx int, src dao.PublishedArticle) domain.Article {
			return c.toDomain(dao.Article(src))
		}), nil
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		//Status:   uint8(art.Status),
		Status: art.Status.ToUint8(),
	}
}

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			// 这里有一个错误
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) { // 预缓存文章
	const size = 1024 * 1024                          // 定义缓存大小
	if len(arts) > 0 && len(arts[0].Content) < size { // 如果文章列表不为空且第一篇文章的内容小于缓存大小
		err := c.cache.Set(ctx, arts[0]) // 设置第一篇文章的缓存
		if err != nil {
			// 记录缓存
		}
	}
}
