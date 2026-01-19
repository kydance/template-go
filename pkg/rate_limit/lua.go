package ratelimit

import "github.com/redis/go-redis/v9"

// allowN implements GCRA (Generic Cell Rate Algorithm) for rate limiting.
// Uses client-provided timestamp to avoid Redis TIME command (non-deterministic).
var allowN = redis.NewScript(`
local rate_limit_key = KEYS[1]
local burst = ARGV[1]
local rate = ARGV[2]
local period = ARGV[3]
local cost = tonumber(ARGV[4])
local now_sec = tonumber(ARGV[5])
local now_microsec = tonumber(ARGV[6])

local emission_interval = period / rate
local increment = emission_interval * cost
local burst_offset = emission_interval * burst

-- Adjust epoch to Jan 1, 2017 to avoid floating point precision issues
local jan_1_2017 = 1483228800
local now = (now_sec - jan_1_2017) + (now_microsec / 1000000)

local tat = redis.call("GET", rate_limit_key)

if not tat then
  tat = now
else
  tat = tonumber(tat)
end

tat = math.max(tat, now)

local new_tat = tat + increment
local allow_at = new_tat - burst_offset

local diff = now - allow_at
local remaining = diff / emission_interval

if remaining < 0 then
  local reset_after = tat - now
  local retry_after = diff * -1
  return {
    0, -- allowed
    0, -- remaining
    tostring(retry_after),
    tostring(reset_after),
  }
end

local reset_after = new_tat - now
if reset_after > 0 then
  redis.call("SET", rate_limit_key, new_tat, "EX", math.ceil(reset_after))
end
local retry_after = -1
return {cost, remaining, tostring(retry_after), tostring(reset_after)}
`)

// allowAtMost is similar to allowN but allows at most n events.
// Uses client-provided timestamp to avoid Redis TIME command (non-deterministic).
var allowAtMost = redis.NewScript(`
local rate_limit_key = KEYS[1]
local burst = ARGV[1]
local rate = ARGV[2]
local period = ARGV[3]
local cost = tonumber(ARGV[4])
local now_sec = tonumber(ARGV[5])
local now_microsec = tonumber(ARGV[6])

local emission_interval = period / rate
local burst_offset = emission_interval * burst

-- Adjust epoch to Jan 1, 2017 to avoid floating point precision issues
local jan_1_2017 = 1483228800
local now = (now_sec - jan_1_2017) + (now_microsec / 1000000)

local tat = redis.call("GET", rate_limit_key)

if not tat then
  tat = now
else
  tat = tonumber(tat)
end

tat = math.max(tat, now)

local diff = now - (tat - burst_offset)
local remaining = diff / emission_interval

if remaining < 1 then
  local reset_after = tat - now
  local retry_after = emission_interval - diff
  return {
    0, -- allowed
    0, -- remaining
    tostring(retry_after),
    tostring(reset_after),
  }
end

if remaining < cost then
  cost = remaining
  remaining = 0
else
  remaining = remaining - cost
end

local increment = emission_interval * cost
local new_tat = tat + increment

local reset_after = new_tat - now
if reset_after > 0 then
  redis.call("SET", rate_limit_key, new_tat, "EX", math.ceil(reset_after))
end

return {
  cost,
  remaining,
  tostring(-1),
  tostring(reset_after),
}
`)
