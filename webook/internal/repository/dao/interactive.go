package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error)
	Get(ctx context.Context, biz string, id int64) (Interactive, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{db: db}
}

// IncrReadCnt 增加阅读计数
func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()

	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		// 在发生冲突时,执行更新操作
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("`read_cnt` + 1"),
			"utime":    now,
		}),
		// 创建
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

// BatchIncrReadCnt 批量增加阅读计数
func (dao *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDao := NewGORMInteractiveDAO(tx) // // 创建一个使用事务的DAO实例
		for i := 0; i < len(bizs); i++ {
			err := txDao.IncrReadCnt(ctx, bizs[i], bizIds[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// InsertLikeInfo 插入点赞信息
func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()

	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Uid:    uid,
			Biz:    biz,
			BizId:  id,
			Status: 1,
			Utime:  now,
		}).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt` + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   id,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error

	})
}

// DeleteLikeInfo 删除点赞信息
func (dao *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context,
	biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).
			Where("uid=? AND biz_id = ? AND biz=?", uid, id, biz).
			Updates(map[string]interface{}{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).
			Where("biz =? AND biz_id=?", biz, id).
			Updates(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt` - 1"),
				"utime":    now,
			}).Error
	})
}

// InsertCollectionBiz 插入收藏信息
func (dao *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Ctime = now
	cb.Utime = now
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&cb).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt` + 1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			Biz:        cb.Biz,
			BizId:      cb.BizId,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
	})
}

// GetLikeInfo 获取点赞信息
func (dao *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz

	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ? AND status = ?",
			biz, id, uid, 1).
		First(&res).Error
	return res, err
}

// GetCollectInfo 获取收藏信息
func (dao *GORMInteractiveDAO) GetCollectInfo(ctx context.Context,
	biz string, id int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ?", biz, id, uid).
		First(&res).Error
	return res, err
}

// Get 获取交互信息
func (dao *GORMInteractiveDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ?", biz, id).
		First(&res).Error
	return res, err
}

// GetByIds 根据ids获取交互信息
func (dao *GORMInteractiveDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	var res []Interactive
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id IN ?", biz, ids).
		First(&res).Error
	return res, err
}

// Interactive 交互信息模型
type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"` // 主键,自增
	// <bizid, biz>
	BizId int64  `gorm:"uniqueIndex:biz_type_id"`                   // 业务ID,与Biz组成唯一索引
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"` // 业务类型,与BizId组成唯一索引

	ReadCnt    int64 // 阅读计数
	LikeCnt    int64 // 点赞计数
	CollectCnt int64 // 收藏计数
	Utime      int64 // 更新时间
	Ctime      int64 // 创建时间
}

// UserCollectionBiz 用户收藏业务模型
type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"` // 主键,自增
	// 这边还是保留了了唯一索引
	Uid   int64  `gorm:"uniqueIndex:uid_biz_type_id"`                   // 用户ID,与BizId和Biz组成唯一索引
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`                   // 业务ID,与Uid和Biz组成唯一索引
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"` // 业务类型,与Uid和BizId组成唯一索引
	// 收藏夹的ID
	// 收藏夹ID本身有索引
	Cid   int64 `gorm:"index"` // 收藏夹ID,有单独的索引
	Utime int64 // 更新时间
	Ctime int64 // 创建时间
}

// UserLikeBiz 用户点赞业务模型
type UserLikeBiz struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`                      // 主键,自增
	Uid    int64  `gorm:"uniqueIndex:uid_biz_type_id"`                   // 用户ID,与BizId和Biz组成唯一索引
	BizId  int64  `gorm:"uniqueIndex:uid_biz_type_id"`                   // 业务ID,与Uid和Biz组成唯一索引
	Biz    string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"` // 业务类型,与Uid和BizId组成唯一索引
	Status int    // 状态
	Utime  int64  // 更新时间
	Ctime  int64  // 创建时间
}
