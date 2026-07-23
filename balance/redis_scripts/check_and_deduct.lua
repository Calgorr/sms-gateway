local balance_key = KEYS[1]
local cost = tonumber(ARGV[1])

local exists = redis.call("EXISTS", balance_key)
if exists == 0 then
    return -1
end

local balance = tonumber(redis.call("GET", balance_key))

if balance < cost then
    return -2
end

local new_balance = redis.call("DECRBY", balance_key, cost)
return new_balance