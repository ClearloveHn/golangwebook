package auth

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/service/sms"
	"github.com/golang-jwt/jwt/v5"
)

// 使用 JWT 来验证和解析模板令牌,确保只有持有有效 JWT 的客户端才能发送短信。

// SMSService 结构体表示短信服务，它包含底层的短信服务 svc 和用于签名和验证 JWT 的密钥 key
type SMSService struct {
	svc sms.Service
	key []byte
}

// Send 方法用于发送短信
func (s *SMSService) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims SMSClaims

	// 使用 jwt.ParseWithClaims 函数解析 tplToken，并将解析结果存储在 claims 中
	_, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		// 提供一个回调函数，用于返回签名和验证 JWT 所需的密钥
		return s.key, nil
	})
	if err != nil {
		return err
	}

	// 调用底层短信服务的 Send 方法发送短信，使用解析后的 claims 中的模板 ID（claims.Tpl）
	return s.svc.Send(ctx, claims.Tpl, args, numbers...)
}

type SMSClaims struct {
	jwt.RegisteredClaims
	Tpl string
	// 额外加字段
}
