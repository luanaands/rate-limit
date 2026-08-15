package configs

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Conf struct {
	ViaCepApiHost          string
	ApiWeatherKey          string
	ApiWeatherHost         string
	RedisAddr              string
	RedisPassword          string
	RedisDB                int
	RateLimitTokenLimit    int
	RateLimitIPLimit       int
	RateLimitBlockDuration time.Duration
}

func LoadConfig(path string) (*Conf, error) {
	v := viper.New()
	v.SetConfigFile(filepath.Join(path, ".env"))
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}

	blockDuration, err := time.ParseDuration(v.GetString("RATE_LIMIT_BLOCK_DURATION"))
	if err != nil {
		return nil, fmt.Errorf("parse RATE_LIMIT_BLOCK_DURATION: %w", err)
	}

	cfg := &Conf{
		ViaCepApiHost:          v.GetString("VIA_CEP_API_HOST"),
		ApiWeatherKey:          v.GetString("API_WEATHER_KEY"),
		ApiWeatherHost:         v.GetString("API_WEATHER_HOST"),
		RedisAddr:              v.GetString("REDIS_ADDR"),
		RedisPassword:          v.GetString("REDIS_PASSWORD"),
		RedisDB:                v.GetInt("REDIS_DB"),
		RateLimitTokenLimit:    v.GetInt("RATE_LIMIT_TOKEN_LIMIT"),
		RateLimitIPLimit:       v.GetInt("RATE_LIMIT_IP_LIMIT"),
		RateLimitBlockDuration: blockDuration,
	}

	return cfg, nil
}
