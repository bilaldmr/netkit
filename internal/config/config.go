package config

import (
	"os"

	"github.com/enteresanlikk/go-toolset/internal/logger"
	"github.com/enteresanlikk/go-toolset/internal/modules/ipcheck"
	"github.com/enteresanlikk/go-toolset/internal/modules/myip"
	"github.com/enteresanlikk/go-toolset/internal/server"
)

type Config struct {
	Log        logger.Config
	HTTPServer server.Config
	IPCheck    ipcheck.Config
	MyIP       myip.Config
}

func Load() Config {
	return Config{
		Log: logger.Config{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
		HTTPServer: server.Config{
			Host: getEnv("HTTP_SERVER_HOST", "0.0.0.0"),
			Port: getEnv("HTTP_SERVER_PORT", "2347"),
		},
		IPCheck: ipcheck.Config{
			IPInfoToken: getEnv("IPINFO_TOKEN", ""),
		},
		MyIP: myip.Config{
			V4URL: getEnv("MYIP_V4_URL", ""),
			V6URL: getEnv("MYIP_V6_URL", ""),
		},
	}
}

func getEnv(key string, fallback string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return fallback
}
