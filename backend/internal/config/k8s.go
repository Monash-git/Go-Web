//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(go-web-mysql:3308)/goweb",
	},
	Redis: RedisConfig{
		Addr: "go-web-redis:6379",
	},
}