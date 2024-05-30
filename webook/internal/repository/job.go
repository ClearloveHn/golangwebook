package repository

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/ClearloveHn/golangwebook/webook/internal/repository/dao"
	"time"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, time time.Time) error
}

type PreemptJobRepository struct {
	dao dao.JobDAO
}

func NewPreemptJobRepository(dao dao.JobDAO) CronJobRepository {
	return &PreemptJobRepository{dao: dao}
}

// Preempt 抢占任务
func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)

	return domain.Job{
		Id:         j.Id,
		Expression: j.Expression,
		Executor:   j.Executor,
		Name:       j.Name,
	}, err
}

// Release 释放任务
func (p *PreemptJobRepository) Release(ctx context.Context, jid int64) error {
	return p.dao.Release(ctx, jid)
}

// UpdateUtime 更新任务的更新时间
func (p *PreemptJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return p.dao.UpdateUtime(ctx, id)
}

// UpdateNextTime 更新任务的下次执行时间
func (p *PreemptJobRepository) UpdateNextTime(ctx context.Context, id int64, time time.Time) error {
	return p.dao.UpdateNextTime(ctx, id, time)
}
