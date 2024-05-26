-- 具体业务
-- 获取传入的键名
local key = KEYS[1]

-- 获取要增加的计数类型(阅读数、点赞数或收藏数)
local cntKey = ARGV[1]
-- 获取增加的计数值
local delta = tonumber(ARGV[2])

-- 检查键是否存在
local exist = redis.call("EXISTS", key)

-- 如果键存在
if exist == 1 then
    -- 使用 HINCRBY 命令对指定字段的值进行增量操作
    redis.call("HINCRBY", key, cntKey, delta)
    -- 返回 1 表示操作成功
    return 1
else
    -- 如果键不存在,返回 0 表示操作失败
    return 0
end