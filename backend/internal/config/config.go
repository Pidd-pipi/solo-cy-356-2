package config

import (
	"github.com/caarlos0/env/v11"
)

// Config 应用配置，全部通过环境变量注入。
type Config struct {
	ServerPort     string `env:"SERVER_PORT" envDefault:"8080"`
	RunMode        string `env:"RUN_MODE" envDefault:"release"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	DBHost         string `env:"DB_HOST" envDefault:"127.0.0.1"`
	DBPort         string `env:"DB_PORT" envDefault:"5432"`
	DBName         string `env:"DB_NAME" envDefault:"communitygarden_db"`
	DBUser         string `env:"DB_USER" envDefault:"communitygarden_user"`
	DBPassword     string `env:"DB_PASSWORD" envDefault:"communitygarden_pwd"`
	DBSSLMode      string `env:"DB_SSLMODE" envDefault:"disable"`
	RedisHost      string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	RedisPort      string `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword  string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB        int    `env:"REDIS_DB" envDefault:"0"`
	JWTSecret      string `env:"JWT_SECRET" envDefault:"change_me_to_a_long_random_string"`
	JWTExpireHours int    `env:"JWT_EXPIRE_HOURS" envDefault:"72"`
}

// Load 从环境变量加载配置。
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
