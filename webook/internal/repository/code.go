package repository

import (
	"context"
	"github.com/ClearloveHn/golangwebook/webook/internal/repository/cache"
)

var ErrCodeVerifyTooMany = cache.ErrCodeVerifyTooMany // 定义错误变量 ErrCodeVerifyTooMany,表示验证码验证次数过多
var ErrCodeSendTooMany = cache.ErrCodeSendTooMany     // 定义错误变量 ErrCodeSendTooMany,表示验证码发送次数过多

//go:generate mockgen -source=./code.go -package=repomocks -destination=./mocks/code.mock.go CodeRepository

// CodeRepository 定义 CodeRepository 接口
type CodeRepository interface {
	Set(ctx context.Context, biz, phone, code string) error            // 设置验证码
	Verify(ctx context.Context, biz, phone, code string) (bool, error) // 验证验证码
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(c cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: c,
	}
}

// Set 实现 CodeRepository 接口的 Set 方法
func (c *CachedCodeRepository) Set(ctx context.Context, biz, phone, code string) error {
	return c.cache.Set(ctx, biz, phone, code) // 调用 CodeCache 的 Set 方法设置验证码
}

// Verify 实现 CodeRepository 接口的 Verify 方法
func (c *CachedCodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return c.cache.Verify(ctx, biz, phone, code) // 调用 CodeCache 的 Verify 方法验证验证码
}
