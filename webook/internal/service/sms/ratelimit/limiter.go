package ratelimit

import (
	"context"
	"errors"
	"github.com/ClearloveHn/golangwebook/webook/internal/service/sms"
)

// 限流器的具体实现可以根据业务需求选择不同的算法和策略，如令牌桶算法、漏桶算法等。限流器的键 key 可以用于区分不同的限流规则
//例如针对不同的短信模板或者不同的发送对象设置不同的限流策略。
//通过使用限流装饰器，可以有效地控制短信发送的速率，避免过于频繁的短信发送对系统造成压力，同时也可以防止短信发送被滥用或者被恶意攻击。

// errLimited 是触发限流时返回的错误
var errLimited = errors.New("触发限流")

// 断言 RateLimitSMSService 实现了 sms.Service 接口
var _ sms.Service = &RateLimitSMSService{}

// RateLimitSMSService 结构体表示带有限流功能的短信服务
type RateLimitSMSService struct {
	svc     sms.Service     // 被装饰的原始短信服务
	limiter limiter.Limiter // 限流器，用于进行速率限制
	key     string          // 限流器的键，用于标识不同的限流规则
}

func NewRateLimitSMSService(svc sms.Service,
	l limiter.Limiter) *RateLimitSMSService {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: l,
		key:     "sms-limiter",
	}
}

// Send 方法用于发送短信，并进行速率限制
func (r *RateLimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 使用限流器进行速率限制
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return err
	}

	// 如果触发限流，返回 errLimited 错误
	if limited {
		return errLimited
	}

	// 如果没有触发限流，调用原始短信服务的 Send 方法发送短信
	return r.svc.Send(ctx, tplId, args, numbers...)
}
