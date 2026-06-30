package config

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type Config struct {
	DBDriver       string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	DBPath         string
	JWTSecret      string
	ServerPort     string
	HolidayAPIURL  string
	AIProvider     string
	AIAPIURL       string
	AIAPIKey       string
	AIModel        string
	BatchSize      int
	MinRestHours   float64
	MaxRetries     int
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DBDriver:      getEnv("DB_DRIVER", "postgres"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "postgres"),
		DBName:        getEnv("DB_NAME", "autoshift"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		DBPath:        getEnv("DB_PATH", "autoshift.db"),
		JWTSecret:     getEnv("JWT_SECRET", "autoshift-secret-change-me"),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		HolidayAPIURL: getEnv("HOLIDAY_API_URL", "https://date.nager.at/api/v3"),
		AIProvider:    getEnv("AI_PROVIDER", "mock"),
		AIAPIURL:      getEnv("AI_API_URL", "https://api.openai.com/v1/chat/completions"),
		AIAPIKey:      getEnv("AI_API_KEY", ""),
		AIModel:       getEnv("AI_MODEL", "gpt-4o"),
		BatchSize:     20,
		MinRestHours:  12.0,
		MaxRetries:    3,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (c *Config) Dialector() gorm.Dialector {
	switch c.DBDriver {
	case "sqlite":
		return sqlite.Open(c.DBPath)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
		return mysql.Open(dsn)
	case "sqlserver":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
		return sqlserver.Open(dsn)
	default:
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
		return postgres.Open(dsn)
	}
}
