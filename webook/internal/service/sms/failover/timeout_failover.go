package failover

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/service/sms"
	"sync/atomic"
)

// TimeoutFailoverSMSService 结构体表示支持超时故障转移的短信服务
type TimeoutFailoverSMSService struct {
	svcs      []sms.Service // 短信服务节点列表
	idx       int32         // 当前正在使用的节点索引
	cnt       int32         // 连续超时的次数
	threshold int32         // 切换阈值，只读
}

// NewTimeoutFailoverSMSService 函数创建一个新的 TimeoutFailoverSMSService 实例
func NewTimeoutFailoverSMSService(svcs []sms.Service, threshold int32) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}

// Send 方法用于发送短信，支持超时故障转移
func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx) // 原子读取当前正在使用的节点索引
	cnt := atomic.LoadInt32(&t.cnt) // 原子读取连续超时的次数

	// 如果连续超时次数达到阈值，尝试切换到下一个节点
	if cnt >= t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs)) // 计算下一个节点的索引
		// 使用 CAS（Compare-And-Swap）操作原子地更新节点索引
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			atomic.StoreInt32(&t.cnt, 0) // 切换成功后，重置连续超时次数为 0
		}
		idx = newIdx // 更新当前使用的节点索引
	}

	svc := t.svcs[idx]                            // 获取当前使用的短信服务节点
	err := svc.Send(ctx, tplId, args, numbers...) // 调用节点的 Send 方法发送短信
	switch err {
	case nil:
		atomic.StoreInt32(&t.cnt, 0) // 发送成功，重置连续超时次数为 0
		return err
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1) // 发送超时，增加连续超时次数
	default:
		// 遇到了其他错误，可以根据需要决定是否增加连续超时次数
		// 如果强调一定是超时错误才切换，那么就不增加
		// 如果是 EOF 之类的错误，你还可以考虑直接切换
	}
	return err
}
