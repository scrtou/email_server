package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
	OAuth2   OAuth2Config
	Frontend FrontendConfig // 新增前端配置
}

type DatabaseConfig struct {
	// DSN string // Kept for potential future use or other DB types
	File string // For SQLite database file path
}

type ServerConfig struct {
	Port string
}

type JWTConfig struct {
	SecretKey string
	ExpiresIn int // 小时
}

type OAuth2Config struct {
	LinuxDo LinuxDoOAuth2Config
}

type LinuxDoOAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

type FrontendConfig struct {
	BaseURL string // 前端基础URL
}

var AppConfig *Config

func Init() {
	// 尝试加载.env文件（在Docker环境中可能不存在，这是正常的）
	if err := godotenv.Load(); err != nil {
		// 只在开发环境中显示警告，生产环境通过环境变量配置
		if os.Getenv("GIN_MODE") != "release" {
			log.Printf("警告: 无法加载.env文件: %v", err)
		}
	}

	AppConfig = &Config{
		Database: DatabaseConfig{
			// DSN: getEnv("DATABASE_URL", "avnadmin:AVNS_icoPVWCDqQgoAM4nCH1@tcp(mysql-yxmysql.c.aivencloud.com:19894)/email-server?charset=utf8mb4&parseTime=True&loc=Local"),
			File: getEnv("SQLITE_FILE", "./gorm.db"), // Default SQLite file path
		},
		Server: ServerConfig{
			Port: "5555", // 固定容器内部端口，外部端口通过BACKEND_PORT配置
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			ExpiresIn: getEnvInt("JWT_EXPIRES_IN", 24),
		},
		OAuth2: OAuth2Config{
			LinuxDo: LinuxDoOAuth2Config{
				ClientID:     getEnv("LINUXDO_CLIENT_ID", ""),
				ClientSecret: getEnv("LINUXDO_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("LINUXDO_REDIRECT_URI", "http://localhost:5555/api/v1/auth/oauth2/linuxdo/callback"),
				AuthURL:      getEnv("LINUXDO_AUTH_URL", "https://connect.linux.do/oauth2/authorize"),
				TokenURL:     getEnv("LINUXDO_TOKEN_URL", "https://connect.linux.do/oauth2/token"),
				UserInfoURL:  getEnv("LINUXDO_USER_INFO_URL", "https://connect.linux.do/api/user"),
			},
		},
		Frontend: FrontendConfig{
			BaseURL: getEnv("FRONTEND_BASE_URL", "http://localhost:8080"),
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
