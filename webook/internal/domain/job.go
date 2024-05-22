package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

// 定义和管理定时任务

// Job 表示一个定时任务
type Job struct {
	Id         int64  // 任务ID,用于唯一标识一个任务
	Name       string // 任务名称,用于描述任务的功能或用途
	Expression string // Cron表达式,用于定义任务的执行时间规则
	Executor   string // 执行器,用于指定执行任务的方法或函数
	Cfg        string // 配置,用于存储任务的额外配置信息,可以是JSON格式的字符串
	CancelFunc func() // 取消函数,用于取消或停止任务的执行
}

// NextTime 方法用于计算任务的下一次执行时间
func (j Job) NextTime() time.Time {
	// 创建一个新的Cron解析器,支持解析秒、分、时、日、月、周等时间单位
	c := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	// 解析任务的Cron表达式
	s, _ := c.Parse(j.Expression)

	// 根据当前时间计算下一次执行时间
	return s.Next(time.Now())
}
