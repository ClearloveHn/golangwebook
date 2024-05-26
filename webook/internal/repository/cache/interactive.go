package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/ClearloveHn/golangwebook/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

const fieldReadCnt = "read_cnt"       // 定义常量 fieldReadCnt,表示阅读数字段名
const fieldLikeCnt = "like_cnt"       // 定义常量 fieldLikeCnt,表示点赞数字段名
const fieldCollectCnt = "collect_cnt" // 定义常量 fieldCollectCnt,表示收藏数字段名

type InteractiveCache interface { // 定义 InteractiveCache 接口,包含交互数据缓存的相关操作
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error // 增加阅读数
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error    // 用于增加点赞数
	DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error    // 减少点赞数
	IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error // 增加收藏数
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client: client,
	}
}

// IncrReadCntIfPresent 定义 IncrReadCntIfPresent 方法,用于增加阅读数
func (i *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := i.key(biz, bizId) // 调用 key 方法生成缓存键

	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Err() // 调用Redis的Eval方法执行Lua脚本,增加阅读数
}

// IncrLikeCntIfPresent 定义 IncrLikeCntIfPresent 方法,用于增加点赞数
func (i *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := i.key(biz, bizId)

	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, 1).Err() // 调用Redis的Eval方法执行Lua脚本,增加点赞数
}

// DecrLikeCntIfPresent 定义 DecrLikeCntIfPresent 方法,用于减少点赞数
func (i *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := i.key(biz, bizId)

	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, -1).Err() // 调用Redis的Eval方法执行Lua脚本,减少点赞数
}

// IncrCollectCntIfPresent 定义 IncrCollectCntIfPresent 方法,用于增加收藏数
func (i *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)

	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldCollectCnt, 1).Err() // 增加收藏数
}

func (i *InteractiveRedisCache) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	key := i.key(biz, id)

	res, err := i.client.HGetAll(ctx, key).Result() // 调用 Redis 的 HGetAll 方法获取指定键的所有字段和值
	if err != nil {
		return domain.Interactive{}, err //
	}

	if len(res) == 0 { // 如果结果为空
		return domain.Interactive{}, ErrKeyNotExist // 返回空的 domain.Interactive 对象和 ErrKeyNotExist 错误
	}

	// 定义 intr 变量,类型为 domain.Interactive
	var intr domain.Interactive

	intr.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64) // 将收藏数字段的值转换为整数,忽略错误
	intr.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)       // 将点赞数字段的值转换为整数,忽略错误
	intr.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)       // 将阅读数字段的值转换为整数,忽略错误

	return intr, nil // 返回 intr 对象和 nil 错误
}

func (i *InteractiveRedisCache) Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error {
	key := i.key(biz, bizId)

	err := i.client.HSet(ctx, key, fieldCollectCnt, res.CollectCnt,
		fieldReadCnt, res.ReadCnt,
		fieldLikeCnt, res.LikeCnt,
	).Err() // 调用 Redis 的 HSet 方法设置交互数据的字段和值

	if err != nil {
		return err
	}

	return i.client.Expire(ctx, key, time.Minute*15).Err() // 调用 Redis 的 Expire 方法设置缓存的过期时间为 15 分钟
}

func (i *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
