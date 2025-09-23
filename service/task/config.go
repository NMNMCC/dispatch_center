package task

import (
	"time"

	"encore.dev/config"
)

type _Config struct {
	WorkerKeepAliveDuration config.String
}

var _cfg = config.Load[*_Config]()

type Config struct {
	WorkerKeepAliveDuration time.Duration `json:"worker_keep_alive_duration"`
}

var cfg = func() *Config {
	workerKeepaliveDuration, err := time.ParseDuration(_cfg.WorkerKeepAliveDuration())
	if err != nil {
		panic(err)
	}

	return &Config{
		WorkerKeepAliveDuration: workerKeepaliveDuration,
	}
}()
