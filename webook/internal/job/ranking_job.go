package job

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"gorm.io/gorm/logger"
	"sync"
	"time"
)

// RankingJob 是一个排名计算任务的结构体
type RankingJob struct {
	svc       service.RankingService // 排名服务,用于执行具体的排名计算逻辑
	l         logger.LoggerV1        // 日志记录器
	timeout   time.Duration          // 任务执行超时时间
	client    *rlock.Client          // Redis 分布式锁客户端
	key       string                 // 分布式锁的键
	localLock *sync.Mutex            // 本地锁,用于保护 lock 字段的并发访问
	lock      *rlock.Lock            // 分布式锁对象
	load      int32                  // 任务的负载,用于统计任务的执行次数
}

func NewRankingJob(
	svc service.RankingService,
	l logger.LoggerV1,
	client *rlock.Client,
	timeout time.Duration) *RankingJob {
	return &RankingJob{svc: svc,
		key:       "job:ranking",
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
		timeout:   timeout}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// Run 方法执行排名计算任务
func (r *RankingJob) Run() error {
	r.localLock.Lock()
	lock := r.lock
	if lock == nil {
		// 如果没有获取到分布式锁,尝试获取分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout,
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
				// 重试的超时
			}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.lock = lock
		r.localLock.Unlock()

		// 在单独的 goroutine 中定期续约分布式锁
		go func() {
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 如果续约失败,释放分布式锁
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}

	// 获取到分布式锁后,执行排名计算任务
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	return r.svc.TopN(ctx)
}

// Close 方法释放分布式锁
func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
