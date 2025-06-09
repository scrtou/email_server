package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
	OAuth2   OAuth2Config
	Frontend FrontendConfig // 新增前端配置
	Backend  BackendConfig
	Security SecurityConfig
}

type SecurityConfig struct {
	EncryptionKey string
}

type BackendConfig struct {
	BaseURL string
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
	LinuxDo   LinuxDoOAuth2Config
	Google    GoogleOAuth2Config
	Microsoft ProviderConfig
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
}

type LinuxDoOAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

type GoogleOAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type FrontendConfig struct {
	BaseURL string // 前端基础URL
}

var AppConfig *Config

func Init() {
	// The run_backend.sh script now handles loading the .env file into environment variables.
	// This function now simply reads from the environment.
	AppConfig = &Config{
		Database: DatabaseConfig{
			File: getEnv("SQLITE_FILE", "./gorm.db"),
		},
		Server: ServerConfig{
			Port: getEnv("BACKEND_PORT", "5555"),
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
			Google: GoogleOAuth2Config{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("GOOGLE_REDIRECT_URI", "http://localhost:5555/api/v1/oauth2/callback/google"),
			},
			Microsoft: ProviderConfig{
				ClientID:     getEnv("MICROSOFT_CLIENT_ID", ""),
				ClientSecret: getEnv("MICROSOFT_CLIENT_SECRET", ""),
			},
		},
		Frontend: FrontendConfig{
			BaseURL: getEnv("FRONTEND_BASE_URL", "http://localhost:8080"),
		},
		Backend: BackendConfig{
			BaseURL: getEnv("BACKEND_BASE_URL", "http://localhost:5555"),
		},
		Security: SecurityConfig{
			EncryptionKey: getEnv("ENCRYPTION_KEY", "12345678901234567890123456789012"), // Must be 32 bytes for AES-256
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
