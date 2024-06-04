package opentelemetry

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/service/sms"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// 这个 OpenTelemetry 装饰器通过在短信发送过程中创建和管理 Span，实现了分布式追踪功能。
//它记录了短信模板 ID、发送短信的事件以及可能出现的错误信息，方便在分布式系统中跟踪和分析短信服务的执行情况。
//通过与其他服务的 Span 建立关联，可以形成完整的分布式追踪链路，帮助定位和解决性能瓶颈和问题。

// Decorator 结构体表示短信服务的 OpenTelemetry 装饰器
type Decorator struct {
	svc    sms.Service  // 原始的短信服务
	tracer trace.Tracer // OpenTelemetry 的 Tracer，用于创建和管理 Span
}

func NewDecorator(svc sms.Service, tracer trace.Tracer) *Decorator {
	return &Decorator{
		svc:    svc,
		tracer: tracer,
	}
}

// Send 方法用于发送短信，并在过程中添加分布式追踪
func (d *Decorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 使用 Tracer 创建一个新的 Span，表示短信发送操作
	ctx, span := d.tracer.Start(ctx, "sms")
	defer span.End()

	// 为 Span 添加属性，记录短信模板 ID
	span.SetAttributes(attribute.String("tplId", tplId))

	// 为 Span 添加事件，表示正在发送短信
	span.AddEvent("发短信")

	// 调用原始短信服务的 Send 方法发送短信
	err := d.svc.Send(ctx, tplId, args)

	// 如果发送短信过程中出现错误，将错误记录到 Span 中
	if err != nil {
		span.RecordError(err)
	}

	return err
}
