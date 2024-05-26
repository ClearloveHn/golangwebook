package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm/logger"
	"strconv"
	"time"
)

// 使用 CronJobBuilder 来构建 cron 任务,并将任务执行时间记录到 Prometheus 摘要向量中

// CronJobBuilder 是一个结构体,用于构建 cron 任务
type CronJobBuilder struct {
	l      logger.LoggerV1        // 日志记录器
	vector *prometheus.SummaryVec // Prometheus 摘要向量,用于记录任务执行时间
}

func NewCronJobBuilder(l logger.LoggerV1, opt prometheus.SummaryOpts) *CronJobBuilder {

	// 创建 Prometheus 摘要向量,用于记录任务执行时间
	vector := prometheus.NewSummaryVec(opt,
		[]string{"job", "success"})
	return &CronJobBuilder{
		l:      l,
		vector: vector}
}

// Build 是 CronJobBuilder 的一个方法,用于构建 cron 任务
func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()

	// 适配器函数
	return cronJobAdapterFunc(func() {
		// 记录任务开始时间
		start := time.Now()
		b.l.Debug("开始运行",
			logger.String("name", name))

		// 执行任务
		err := job.Run()
		if err != nil {
			b.l.Error("执行失败",
				logger.Error(err),
				logger.String("name", name))
		}

		b.l.Debug("结束运行",
			logger.String("name", name))

		// 计算任务执行时间并记录到 Prometheus 摘要向量
		duration := time.Since(start)
		b.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).
			Observe(float64(duration.Milliseconds()))
	})
}

// cronJobAdapterFunc 是一个函数类型,用于适配 cron 任务的执行
type cronJobAdapterFunc func()

// Run 是 cronJobAdapterFunc 的一个方法,用于执行 cron 任务
func (c cronJobAdapterFunc) Run() {
	c()
}
