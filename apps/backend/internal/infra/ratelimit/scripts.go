package ratelimit

// tokenBucketScript is the Lua script for atomic token bucket rate limiting.
// KEYS[1] = bucket key
// ARGV[1] = capacity (burst)
// ARGV[2] = refill_rate (tokens per interval)
// ARGV[3] = refill_interval (microseconds)
// ARGV[4] = now (current time in microseconds)
// Returns: {allowed (1/0), remaining_tokens}
const tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local refill_interval = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(bucket[1])
local last_refill = tonumber(bucket[2])

if tokens == nil then
    tokens = capacity
    last_refill = now
end

local elapsed = now - last_refill
local refills = math.floor(elapsed / refill_interval)
if refills > 0 then
    tokens = math.min(capacity, tokens + refills * refill_rate)
    last_refill = last_refill + refills * refill_interval
end

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call('HMSET', key, 'tokens', tokens, 'last_refill', last_refill)
redis.call('EXPIRE', key, math.ceil(capacity / refill_rate) * refill_interval / 1000000 + 60)

return {allowed, tokens}
`

// slidingWindowScript is the Lua script for sliding window rate limiting.
// KEYS[1] = window key
// ARGV[1] = max requests in window
// ARGV[2] = window size in seconds
// ARGV[3] = current timestamp (seconds)
// Returns: {allowed (1/0), remaining}
const slidingWindowScript = `
local key = KEYS[1]
local max = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local window_start = now - window
redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

local count = redis.call('ZCARD', key)
local allowed = 0
local remaining = max - count

if count < max then
    redis.call('ZADD', key, now, now .. ':' .. math.random(1000000))
    allowed = 1
    remaining = remaining - 1
end

redis.call('EXPIRE', key, window + 1)

return {allowed, remaining}
`
