package database

import (
	"log"

	"email_server/config"
	"email_server/models"
	"email_server/utils"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}

	log.Println("✅ 数据库连接成功")

	// 自动迁移数据表
	err = DB.AutoMigrate(
		&models.User{},
		&models.EmailAccount{},
		&models.Platform{},
		&models.PlatformRegistration{},
		&models.ServiceSubscription{},
		&models.OAuthProvider{},
		&models.UserOAuthToken{},
		&models.OAuth2State{},
	)
	if err != nil {
		log.Fatal("❌ 数据库表自动迁移失败:", err)
	}
	log.Println("🎉 数据表自动迁移完成")

	// 创建默认管理员账户
	createDefaultAdminUser()

	// 播种OAuth提供商
	seedOAuthProviders()
}

func createDefaultAdminUser() {
	var adminUser models.User
	err := DB.Where("username = ?", "admin").First(&adminUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			hashedPassword, hashErr := utils.HashPassword("password")
			if hashErr != nil {
				log.Fatalf("❌ 创建默认管理员时密码哈希失败: %v", hashErr)
				return
			}
			defaultAdmin := models.User{
				Username: "admin",
				Email:    "admin@example.com",
				Password: hashedPassword,
				Role:     models.RoleAdmin,
				Status:   models.StatusActive,
			}
			if createErr := DB.Create(&defaultAdmin).Error; createErr != nil {
				log.Fatalf("❌ 创建默认管理员账户失败: %v", createErr)
			} else {
				log.Println("🔑 默认管理员账户 'admin' 创建成功 (密码: password)")
			}
		} else {
			log.Fatalf("❌ 查询默认管理员账户失败: %v", err)
		}
	} else {
		log.Println("ℹ️ 默认管理员账户 'admin' 已存在")
	}
}

// seedOAuthProviders 在数据库中创建或更新OAuth提供商的配置
func seedOAuthProviders() {
	// 从配置中读取提供商信息
	providers := map[string]config.ProviderConfig{
		"google": {
			ClientID:     config.AppConfig.OAuth2.Google.ClientID,
			ClientSecret: config.AppConfig.OAuth2.Google.ClientSecret,
		},
		"microsoft": {
			ClientID:     config.AppConfig.OAuth2.Microsoft.ClientID,
			ClientSecret: config.AppConfig.OAuth2.Microsoft.ClientSecret,
		},
	}

	staticData := map[string]models.OAuthProvider{
		"google": {
			AuthURL:    "https://accounts.google.com/o/oauth2/auth",
			TokenURL:   "https://oauth2.googleapis.com/token",
			Scopes:     "https://mail.google.com/,https://www.googleapis.com/auth/userinfo.email,https://www.googleapis.com/auth/userinfo.profile",
			IMAPServer: "imap.gmail.com",
			IMAPPort:   993,
		},
		"microsoft": {
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			//Scopes:     "offline_access,User.Read,Mail.Send,IMAP.AccessAsUser.All",
			Scopes:     "offline_access User.Read Mail.ReadWrite",
			IMAPServer: "outlook.office365.com",
			IMAPPort:   993,
		},
	}

	for name, config := range providers {
		if config.ClientID == "" || config.ClientSecret == "" {
			log.Printf("ℹ️ 跳过为 '%s' 播种，因为未在配置文件中提供 ClientID 或 ClientSecret。", name)
			continue
		}

		encryptedSecret, err := utils.Encrypt([]byte(config.ClientSecret))
		if err != nil {
			log.Printf("❌ 加密 '%s' 的Client Secret失败: %v", name, err)
			continue
		}

		providerRecord := models.OAuthProvider{
			Name:                  name,
			ClientID:              config.ClientID,
			ClientSecretEncrypted: encryptedSecret,
			AuthURL:               staticData[name].AuthURL,
			TokenURL:              staticData[name].TokenURL,
			Scopes:                staticData[name].Scopes,
			IMAPServer:            staticData[name].IMAPServer,
			IMAPPort:              staticData[name].IMAPPort,
		}

		// 使用 Assign 和 FirstOrCreate 来创建或更新记录
		if err := DB.Where(models.OAuthProvider{Name: name}).Assign(providerRecord).FirstOrCreate(&models.OAuthProvider{}).Error; err != nil {
			log.Printf("❌ 播种OAuth提供商 '%s' 失败: %v", name, err)
		} else {
			log.Printf("🌱 OAuth提供商 '%s' 已成功配置。", name)
		}
	}
}
