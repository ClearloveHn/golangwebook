package tencent

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/zap"
)

// Service 结构体表示腾讯云短信服务
type Service struct {
	client   *sms.Client // 腾讯云短信客户端
	appId    *string     // 短信应用 ID
	signName *string     // 短信签名
}

func NewService(client *sms.Client, appId string, signName string) *Service {
	return &Service{
		client:   client,
		appId:    &appId,
		signName: &signName,
	}
}

// Send 方法用于发送短信
func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 创建发送短信的请求
	request := sms.NewSendSmsRequest()
	request.SetContext(ctx)
	request.SmsSdkAppId = s.appId
	request.SignName = s.signName
	request.TemplateId = ekit.ToPtr[string](tplId)
	request.TemplateParamSet = s.toPtrSlice(args)
	request.PhoneNumberSet = s.toPtrSlice(numbers)

	// 发送短信请求
	response, err := s.client.SendSms(request)

	// 记录请求和响应的日志
	zap.L().Debug("请求腾讯SendSMS接口",
		zap.Any("req", request),
		zap.Any("resp", response))

	// 处理异常
	if err != nil {
		return err
	}

	// 遍历发送状态集合
	for _, statusPtr := range response.Response.SendStatusSet {
		if statusPtr == nil {
			// 不可能进来这里
			continue
		}

		status := *statusPtr
		if status.Code == nil || *(status.Code) != "Ok" {
			// 发送失败
			return fmt.Errorf("发送短信失败 code: %s, msg: %s", *status.Code, *status.Message)
		}
	}

	// 发送成功
	return nil
}

// toPtrSlice 方法将字符串切片转换为指针切片
func (s *Service) toPtrSlice(data []string) []*string {
	return slice.Map[string, *string](data,
		func(idx int, src string) *string {
			return &src
		})
}
