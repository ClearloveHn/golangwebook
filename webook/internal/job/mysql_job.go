package job

import (
	"context"
	"fmt"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"golang.org/x/sync/semaphore"
	"gorm.io/gorm/logger"
	"time"
)

// 实现了一个任务调度器 Scheduler,用于调度和执行定时任务

// Executor 是一个接口,表示任务执行器
// 任务执行器负责执行具体的任务逻辑
type Executor interface {
	// Name 方法返回执行器的名称
	Name() string

	// Exec 方法执行任务
	Exec(ctx context.Context, j domain.Job) error
}

// LocalFuncExecutor 调用本地方法的
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, j domain.Job) error)}
}

// Name 方法返回执行器的名称
func (l *LocalFuncExecutor) Name() string {
	return "local"
}

// RegisterFunc 方法用于注册本地函数到 LocalFuncExecutor
func (l *LocalFuncExecutor) RegisterFunc(name string, f func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = f
}

// Exec 方法执行任务
func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	// 根据任务的名称获取对应的本地函数
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未注册本地方法 %s", j.Name)
	}

	return fn(ctx, j)
}

// Scheduler 是一个任务调度器
type Scheduler struct {
	dbTimeout time.Duration // 数据库操作超时时间

	svc service.CronJobService // 任务服务,用于获取和更新任务状态

	executors map[string]Executor // 执行器映射,用于存储不同类型的执行器
	l         logger.LoggerV1     // 日志记录器

	limiter *semaphore.Weighted // 信号量,用于限制并发执行的任务数量
}

func NewScheduler(svc service.CronJobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{
		svc:       svc,
		dbTimeout: time.Second,
		limiter:   semaphore.NewWeighted(100), // 创建一个权重为100的信号量,表示最多可以并发执行100个任务
		l:         l,
		executors: map[string]Executor{},
	}
}

// RegisterExecutor 方法用于注册执行器到 Scheduler
func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

// Schedule 方法是 Scheduler 的主要调度逻辑
func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		// 检查 ctx 是否已经被取消或超时
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 尝试获取一个信号量,如果没有可用的信号量,则会阻塞等待
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		// 从任务服务中获取一个待执行的任务
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 有 Error
			// 最简单的做法就是直接下一轮，也可以睡一段时间
			continue
		}

		// 根据任务的执行器类型获取对应的执行器
		exec, ok := s.executors[j.Executor]
		if !ok {
			// 如果找不到对应的执行器,可以直接中断调度,也可以进入下一次循环
			s.l.Error("找不到执行器",
				logger.Int64("jid", j.Id),
				logger.String("executor", j.Executor))
			continue
		}

		// 在单独的 goroutine 中执行任务
		go func() {
			defer func() {
				s.limiter.Release(1)
				// 这边要释放掉
				j.CancelFunc()
			}()

			// 使用对应的执行器执行任务
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				s.l.Error("执行任务失败",
					logger.Int64("jid", j.Id),
					logger.Error(err1))
				return
			}

			// 执行完成后,重置任务的下次执行时间
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("重置下次执行时间失败",
					logger.Int64("jid", j.Id),
					logger.Error(err1))
			}
		}()
	}
}
