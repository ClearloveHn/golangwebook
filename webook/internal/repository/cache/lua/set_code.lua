-- 获取传入的键名
local key = KEYS[1]
-- 构建计数器的键名
local cntKey = key..":cnt"

-- 获取准备存储的验证码值
local val = ARGV[1]

-- 获取键的剩余过期时间(TTL)
local ttl = tonumber(redis.call("ttl", key))

-- 如果键存在但没有设置过期时间
if ttl == -1 then
    -- 返回 -2,表示键已存在但没有过期时间
    return -2
-- 如果键不存在或者剩余过期时间小于540秒
elseif ttl == -2 or ttl < 540 then
    -- 可以发送验证码
    -- 设置验证码的值
    redis.call("set", key, val)
    -- 设置验证码的过期时间为600秒(10分钟)
    redis.call("expire", key, 600)
    -- 设置计数器的初始值为3
    redis.call("set", cntKey, 3)
    -- 设置计数器的过期时间为600秒(10分钟)
    redis.call("expire", cntKey, 600)
    -- 返回0,表示可以发送验证码
    return 0
else
    -- 发送频率太高
    -- 返回-1,表示发送频率过高,不允许发送验证码
    return -1
end