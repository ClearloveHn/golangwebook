package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string // 内嵌的 Lua 脚本，用于设置验证码
	//go:embed lua/verify_code.lua
	luaVerifyCode string // 内嵌的 Lua 脚本，用于验证验证码

	ErrCodeSendTooMany   = errors.New("发送太频繁") // 发送验证码过于频繁的错误
	ErrCodeVerifyTooMany = errors.New("验证太频繁") // 验证验证码过于频繁的错误
)

//go:generate mockgen -source=./code.go -package=cachemocks -destination=./mocks/code.mock.go CodeCache

// CodeCache 是一个接口，定义了验证码缓存的相关操作
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error            // 设置验证码
	Verify(ctx context.Context, biz, phone, code string) (bool, error) // 验证验证码
}

type RedisCodeCache struct {
	cmd redis.Cmdable // Redis 命令执行器
}

func NewCodeCache(cmd redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		cmd: cmd,
	}
}

// Set 方法用于设置验证码
func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	// 调用 Redis 的 Eval 方法执行 Lua 脚本，传入验证码的键和验证码值作为参数
	res, err := c.cmd.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}

	// 根据 Lua 脚本的返回值进行错误处理
	switch res {
	case -2:
		return errors.New("验证码已存在，但是没有过期时间")
	case -1:
		return ErrCodeSendTooMany
	default:
		return nil
	}
}

// Verify 方法用于验证验证码
func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	// 调用 Redis 的 Eval 方法执行 Lua 脚本，传入验证码的键和验证码值作为参数
	res, err := c.cmd.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		// 调用 Redis 出错
		return false, err
	}

	// 根据 Lua 脚本的返回值进行结果处理
	switch res {
	case -2:
		return false, nil
	case -1:
		return false, ErrCodeVerifyTooMany
	default:
		return true, nil
	}
}

// key 方法用于生成验证码的缓存键
func (c *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
