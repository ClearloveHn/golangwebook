package prometheus

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/ClearloveHn/golangwebook/webook/internal/service/oauth2/wechat"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// Decorator 结构体是一个 wechat.Service 的装饰器，用于添加 Prometheus 监控
type Decorator struct {
	wechat.Service                    // 内嵌 wechat.Service 接口，实现装饰器模式
	sum            prometheus.Summary // Prometheus Summary 类型的度量指标，用于记录请求的耗时
}

// NewDecorator 函数创建一个新的 Decorator 实例
func NewDecorator(svc wechat.Service, sum prometheus.Summary) *Decorator {
	return &Decorator{
		Service: svc,
		sum:     sum,
	}
}

// VerifyCode 方法对 wechat.Service 的 VerifyCode 方法进行装饰，添加 Prometheus 监控
func (d *Decorator) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	start := time.Now() // 记录开始时间
	defer func() {
		duration := time.Since(start).Milliseconds() // 计算方法执行耗时，单位为毫秒
		d.sum.Observe(float64(duration))             // 将耗时记录到 Prometheus Summary 度量指标中
	}()
	return d.Service.VerifyCode(ctx, code) // 调用原始的 VerifyCode 方法
}
