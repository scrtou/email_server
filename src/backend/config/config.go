package config

import (
    "os"
    "strconv"
)

type Config struct {
    Database DatabaseConfig
    Server   ServerConfig
    JWT      JWTConfig
}

type DatabaseConfig struct {
    DSN string
}

type ServerConfig struct {
    Port string
}

type JWTConfig struct {
    SecretKey string
    ExpiresIn int // 小时
}

var AppConfig *Config

func Init() {
    AppConfig = &Config{
        Database: DatabaseConfig{
            DSN: getEnv("DATABASE_URL", "avnadmin:AVNS_icoPVWCDqQgoAM4nCH1@tcp(mysql-yxmysql.c.aivencloud.com:19894)/email-server?charset=utf8mb4&parseTime=True&loc=Local"),
        },
        Server: ServerConfig{
            Port: getEnv("SERVER_PORT", "5555"),
        },
        JWT: JWTConfig{
            SecretKey: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
            ExpiresIn: getEnvInt("JWT_EXPIRES_IN", 24),
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
