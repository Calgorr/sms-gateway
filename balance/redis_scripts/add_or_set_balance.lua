local current = redis.call("GET", KEYS[1])
if current then
	return redis.call("INCRBY", KEYS[1], ARGV[1])
else
	redis.call("SET", KEYS[1], ARGV[1])
	return tonumber(ARGV[1])
end