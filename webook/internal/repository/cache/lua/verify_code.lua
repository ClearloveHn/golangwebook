-- 获取传入的键名
local key = KEYS[1]
-- 构建计数器的键名
local cntKey = key..":cnt"

-- 获取用户输入的验证码
local expectedCode = ARGV[1]

-- 获取当前验证码的剩余验证次数
local cnt = tonumber(redis.call("get", cntKey))
-- 获取存储的验证码
local code = redis.call("get", key)

-- 如果剩余验证次数为空或者小于等于0
if cnt == nil or cnt <= 0 then
    -- 验证次数已耗尽
    -- 返回-1，表示验证次数耗尽
    return -1
end

-- 如果存储的验证码与用户输入的验证码相等
if code == expectedCode then
    -- 验证码匹配成功
    -- 将剩余验证次数设置为0
    redis.call("set", cntKey, 0)
    -- 返回0，表示验证码匹配成功
    return 0
else
    -- 验证码不匹配
    -- 使用 DECR 命令将剩余验证次数减1
    redis.call("decr", cntKey)
    -- 返回-2，表示验证码不匹配
    return -2
end