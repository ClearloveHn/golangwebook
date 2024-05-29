package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGORMJobDAO(db *gorm.DB) JobDAO {
	return &GORMJobDAO{db: db}
}

// Preempt 抢占一个Job
func (dao *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := dao.db.WithContext(ctx)

	for {
		var j Job
		now := time.Now().UnixMilli()

		// 查找状态为等待中且下次执行时间小于当前时间的Job
		err := db.Where("status = ? AND next_time < ?", jobStatusWaiting, now).First(&j).Error
		if err != nil {
			return j, err
		}

		// 抢占
		res := db.WithContext(ctx).Model(&Job{}).Where("id = ? AND version = ?", j.Id, j.Version).
			Updates(map[string]any{
				"status":  jobStatusRunning,
				"version": j.Version + 1,
				"utime":   now,
			})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			continue
		}

		return j, nil
	}
}

// Release 释放一个job
func (dao *GORMJobDAO) Release(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  now,
	}).Error
}

// UpdateUtime 更新job的更新时间
func (dao *GORMJobDAO) UpdateUtime(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).Updates(map[string]any{
		"utime": now,
	}).Error
}

// UpdateNextTime 更新job的下次执行时间
func (dao *GORMJobDAO) UpdateNextTime(ctx context.Context, jid int64, t time.Time) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).Updates(map[string]any{
		"utime":     now,
		"next_time": t.UnixMilli(),
	}).Error
}

// Job 作业模型
type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"` // 主键,自增
	Name       string `gorm:"type:varchar(128);unique"` // 名称,唯一
	Executor   string // 执行器
	Expression string // 表达式
	Cfg        string // 配置
	// 状态来表达,是不是可以抢占,有没有被人抢占
	Status int // 状态

	Version int // 版本

	NextTime int64 `gorm:"index"` // 下次执行时间,有索引

	Utime int64 // 更新时间
	Ctime int64 // 创建时间
}

const (
	// jobStatusWaiting 没人抢
	jobStatusWaiting = iota
	// jobStatusRunning 已经被人抢了
	jobStatusRunning
	// jobStatusPaused 不再需要调度了
	jobStatusPaused
)
