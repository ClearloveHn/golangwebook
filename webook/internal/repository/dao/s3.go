package dao

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

type ArticleS3DAO struct {
	ArticleGORMDAO
	oss *s3.S3 // AWS S3 客户端实例
}

func NewArticleS3DAO(db *gorm.DB, oss *s3.S3) *ArticleS3DAO {
	return &ArticleS3DAO{ArticleGORMDAO: ArticleGORMDAO{db: db}, oss: oss}
}

// Sync 同步文章,如果 art.Id > 0 则更新,否则插入
func (a *ArticleS3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id

	// 开启数据库事务
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			err error
		)

		// 创建一个新的 ArticleGORMDAO 实例,使用事务中的 tx
		dao := NewArticleGORMDAO(tx)

		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}

		// 更新or创建PublishedArticleV2
		art.Id = id
		now := time.Now().UnixMilli()
		pubArt := PublishedArticleV2{
			Id:       art.Id,
			Title:    art.Title,
			AuthorId: art.AuthorId,
			Ctime:    now,
			Utime:    now,
			Status:   art.Status,
		}
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  pubArt.Title,
				"utime":  now,
				"status": pubArt.Status,
			}),
		}).Create(&pubArt).Error
		return err
	})
	if err != nil {
		return 0, err
	}

	// 上传文章内容到 S3
	_, err = a.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("webook-1314583317"),           // S3 存储桶名称
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)), // 对象键,使用文章 ID 作为键
		Body:        bytes.NewReader([]byte(art.Content)),              // 文章内容
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),    // 内容类型
	})

	return id, err
}

// SyncStatus 同步文章状态
func (a *ArticleS3DAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	now := time.Now().UnixMilli()

	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? and author_id = ?", uid, id).
			Updates(map[string]any{
				"utime": now,
				"stime": status,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("ID 不对或者创作者不对")
		}
		return tx.Model(&PublishedArticleV2{}).
			Where("id = ?", uid).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			}).Error
	})
	if err != nil {
		return err
	}

	const statusPrivate = 3
	if status == statusPrivate {
		// 如果状态为私有,删除 S3 中对应的对象
		_, err = a.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string]("webook-1314583317"),       // S3 存储桶名称
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)), // 对象键,使用文章 ID 作为键
		})
	}

	return err
}

type PublishedArticleV2 struct {
	Id    int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	// 我要根据创作者ID来查询
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}
