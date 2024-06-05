package service

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/ClearloveHn/golangwebook/webook/internal/repository"
	"gorm.io/gorm/logger"
	"time"
)

// CronJobService 接口定义了定时任务服务的方法
type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.LoggerV1
	refreshInterval time.Duration
}

func NewCronJobService(repo repository.CronJobRepository, l logger.LoggerV1) CronJobService {
	return &cronJobService{repo: repo,
		l:               l,
		refreshInterval: time.Minute}
}

// Preempt 方法用于抢占一个定时任务
func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}

	// 创建一个定时器，用于定期刷新任务的更新时间
	ticker := time.NewTicker(c.refreshInterval)

	go func() {
		for range ticker.C {
			c.refresh(j.Id) // 在新的 goroutine 中定期刷新任务的更新时间
		}
	}()

	j.CancelFunc = func() {
		ticker.Stop() // 停止定时器
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := c.repo.Release(ctx, j.Id) // 释放任务
		if err != nil {
			c.l.Error("释放 job 失败",
				logger.Error(err),
				logger.Int64("jib", j.Id))
		}
	}
	return j, err
}

// ResetNextTime 方法用于重置定时任务的下次执行时间
func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.UpdateNextTime(ctx, j.Id, nextTime)
}

// refresh 方法用于刷新定时任务的更新时间
func (c *cronJobService) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		c.l.Error("续约失败", logger.Error(err),
			logger.Int64("jid", id))
	}
}
