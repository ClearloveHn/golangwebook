package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"net/http"
	"net/url"
)

// Service 接口定义了微信登录服务需要实现的方法
type Service interface {
	// AuthURL 方法用于生成微信登录的授权URL
	AuthURL(ctx context.Context, state string) (string, error)

	// VerifyCode 方法用于验证微信返回的授权码，并获取用户信息
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

// redirectURL 是微信登录回调的URL，需要进行URL编码
var redirectURL = url.PathEscape("https://meoying.com/oauth2/wechat/callback")

// service 结构体实现了 Service 接口
type service struct {
	appID     string          // 微信应用的 App ID
	appSecret string          // 微信应用的 App Secret
	client    *http.Client    // HTTP 客户端，用于发送请求
	l         logger.LoggerV1 // 日志记录器
}

func NewService(appID string, appSecret string, l logger.LoggerV1) Service {
	return &service{
		appID:     appID,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

// AuthURL 方法生成微信登录的授权URL
func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	// 微信登录授权URL的模板
	const authURLPattern = `https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect`
	// 使用 fmt.Sprintf 函数将参数填充到模板中，生成最终的URL
	return fmt.Sprintf(authURLPattern, s.appID, redirectURL, state), nil
}

// VerifyCode 方法验证微信返回的授权码，并获取用户信息
func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	// 构建获取访问令牌的URL
	accessTokenUrl := fmt.Sprintf(`https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code`,
		s.appID, s.appSecret, code)

	// 创建一个带有上下文的 HTTP GET 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenUrl, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	// 发送 HTTP 请求
	httpResp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	// 解析微信接口返回的 JSON 数据
	var res Result
	err = json.NewDecoder(httpResp.Body).Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	// 检查微信接口返回的错误码
	if res.ErrCode != 0 {
		return domain.WechatInfo{},
			fmt.Errorf("调用微信接口失败 errcode %d, errmsg %s", res.ErrCode, res.ErrMsg)
	}
	// 返回获取到的用户信息
	return domain.WechatInfo{
		UnionId: res.UnionId,
		OpenId:  res.OpenId,
	}, nil
}

// Result 结构体表示微信接口返回的 JSON 数据
type Result struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	ExpiresIn    int64  `json:"expires_in"`    // 访问令牌的过期时间，单位为秒
	RefreshToken string `json:"refresh_token"` // 刷新访问令牌的令牌
	OpenId       string `json:"openid"`        // 用户的唯一标识
	Scope        string `json:"scope"`         // 用户授权的作用域，多个作用域用逗号分隔
	UnionId      string `json:"unionid"`       // 用户在开放平台的唯一标识符，在满足 UnionID 下发条件的情况下会返回

	ErrCode int    `json:"errcode"` // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误信息
}
