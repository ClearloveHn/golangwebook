package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

//go:generate mockgen -source=./article.go -package=daomocks -destination=./mocks/article.mock.go ArticleDAO

// ArticleDAO 接口,包含文章相关的数据访问操作
type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)                                          // 插入文章
	UpdateById(ctx context.Context, entity Article) error                                            // 根据 ID 更新文章
	Sync(ctx context.Context, entity Article) (int64, error)                                         // 同步文章
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error                         // 同步文章状态
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)            // 根据作者 ID 获取文章列表
	GetById(ctx context.Context, id int64) (Article, error)                                          // 根据 ID 获取文章
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)                              // 根据 ID 获取已发布的文章
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) // 获取已发布的文章列表
}

// Article 结构体,表示文章实体
type Article struct {
	Id       int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"` // 文章 ID,主键,自增
	Title    string `gorm:"type=varchar(4096)" bson:"title,omitempty"`    // 文章标题
	Content  string `gorm:"type=BLOB" bson:"content,omitempty"`           // 文章内容
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`             // 作者 ID,建立索引
	Status   uint8  `bson:"status,omitempty"`                             // 文章状态
	Ctime    int64  `bson:"ctime,omitempty"`                              // 创建时间
	Utime    int64  `bson:"utime,omitempty"`                              // 更新时间
}

// PublishedArticle 类型,表示已发布的文章,与 Article 结构相同
type PublishedArticle Article

type ArticleGORMDAO struct {
	db *gorm.DB
}

func NewArticleGORMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{
		db: db,
	}
}

// Insert 方法,插入文章
func (a *ArticleGORMDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli() // 获取当前时间的毫秒时间戳
	art.Ctime = now               // 设置文章的创建时间
	art.Utime = now               // 设置文章的更新时间

	// 使用 GORM 的 Create 方法插入文章
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// UpdateById 方法,根据 ID 更新文章
func (a *ArticleGORMDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()

	// 使用 GORM 的 Model 方法更新文章
	res := a.db.WithContext(ctx).Model(&art).Where("id=? AND author_id=?", art.Id, art.AuthorId).Updates(map[string]interface{}{
		"title":   art.Title,   // 更新文章标题
		"content": art.Content, // 更新文章内容
		"status":  art.Status,  // 更新文章状态
		"utime":   now,         // 更新文章的更新时间
	})
	if res.Error != nil {
		return res.Error
	}

	// 如果没有更新任何行
	if res.RowsAffected == 0 {
		return errors.New("ID不对或创作者不对") // 返回自定义错误
	}

	return nil
}

// Sync 方法,同步文章
func (a *ArticleGORMDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id // 获取文章 ID

	// 使用 GORM 的 Transaction 方法开启事务
	// tx数据库实例
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			err error // 定义错误变量
		)

		// 创建一个新的 ArticleGORMDAO 实例,使用事务中的数据库实例
		dao := NewArticleGORMDAO(tx)

		// 如果文章 ID 大于 0,表示更新文章
		if id > 0 {
			err = dao.UpdateById(ctx, art) // 调用 UpdateById 方法更新文章
		} else { // 如果文章 ID 小于等于 0,表示插入文章
			id, err = dao.Insert(ctx, art) // 调用 Insert 方法插入文章,并获取插入的文章 ID
		}
		if err != nil { // 如果更新或插入出错
			return err // 返回错误,事务将会回滚
		}

		// 已发布文章操作
		art.Id = id                     // 更新文章 ID
		now := time.Now().UnixMilli()   // 获取当前时间的毫秒时间戳
		pubArt := PublishedArticle(art) // 将文章转换为已发布的文章
		pubArt.Ctime = now              // 设置已发布文章的创建时间
		pubArt.Utime = now              // 设置已发布文章的更新时间

		// 使用 GORM 的 Clauses 方法构建 ON CONFLICT 语句
		// ON CONFLICT 是 SQL 中的一个子句,用于处理插入数据时的冲突情况,表示当插入的数据与已有数据的 ID 冲突时,执行更新操作而不是插入操作。
		// 如果文章的 ID 已经存在,则会触发 ON CONFLICT 子句,执行更新操作;否则,会执行插入操作。
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}}, // 指定 ON CONFLICT 的列为 ID 列
			// 指定发生冲突时要更新的字段和值
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   pubArt.Title,   // 更新文章标题
				"content": pubArt.Content, // 更新文章内容
				"utime":   now,            // 更新文章的更新时间
				"status":  pubArt.Status,  // 更新文章状态
			}),
			// 没有触发ON CONFLICT 正常创建
		}).Create(&pubArt).Error // 使用 GORM 的 Create 方法插入或更新已发布的文章

		return err
	})

	return id, err
}

// SyncStatus 方法,同步文章状态
func (a *ArticleGORMDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	now := time.Now().UnixMilli()

	// 使用 GORM 的 Transaction 方法开启事务
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}). // 使用 GORM 的 Model 方法更新文章状态
						Where("id = ? and author_id = ?", uid, id). // 指定更新条件
						Updates(map[string]any{                     // 指定更新内容
				"utime":  now,    // 更新文章的更新时间
				"status": status, // 更新文章状态
			})
		if res.Error != nil {
			return res.Error // 返回错误,事务将会回滚
		}
		if res.RowsAffected != 1 { // 如果更新的行数不等于 1
			return errors.New("ID 不对或者创作者不对") // 返回自定义错误,事务将会回滚
		}

		// 使用 GORM 的 Model 方法更新已发布文章的状态
		return tx.Model(&PublishedArticle{}).
			Where("id = ?", uid).   // 指定更新条件
			Updates(map[string]any{ // 指定更新内容
				"utime":  now,    // 更新已发布文章的更新时间
				"status": status, // 更新已发布文章的状态
			}).Error // 返回可能的错误
	})
}

// GetByAuthor 方法,根据作者 ID 获取文章列表
func (a *ArticleGORMDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article // 定义文章切片

	err := a.db.WithContext(ctx). // 使用 GORM 的 DB 方法查询文章
					Where("author_id = ?", uid). // 指定查询条件
					Offset(offset).Limit(limit). // 指定偏移量和限制数量  (实现分页查询)
					Order("utime DESC").         // 按照更新时间倒序排序
					Find(&arts).Error            // 查询文章列表

	return arts, err
}

// GetById 方法,根据 ID 获取文章
func (a *ArticleGORMDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article // 定义文章变量

	err := a.db.WithContext(ctx). // 使用 GORM 的 DB 方法查询文章
					Where("id = ?", id). // 指定查询条件
					First(&art).Error    // 查询文章

	return art, err
}

// GetPubById 方法,根据 ID 获取已发布的文章
func (a *ArticleGORMDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var res PublishedArticle      // 定义已发布文章变量
	err := a.db.WithContext(ctx). // 使用 GORM 的 DB 方法查询已发布文章
					Where("id = ?", id). // 指定查询条件
					First(&res).Error    // 查询已发布文章

	return res, err
}

// ListPub 方法,获取已发布的文章列表
func (a *ArticleGORMDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var res []PublishedArticle       // 定义已发布文章切片
	const ArticleStatusPublished = 2 // 定义已发布文章的状态常量

	err := a.db.WithContext(ctx). // 使用 GORM 的 DB 方法查询已发布文章
					Where("utime < ? AND status = ?", // 指定查询条件
						start.UnixMilli(), ArticleStatusPublished). // 查询更新时间小于指定时间且状态为已发布的文章
		Offset(offset).Limit(limit). // 指定偏移量和限制数量
		First(&res).Error            // 查询已发布文章列表

	return res, err
}
