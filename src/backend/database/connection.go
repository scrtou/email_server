package database

import (
	"log"
	// "email_server/config" // Config will be handled differently or passed to Init
	"email_server/models" // Assuming models will be defined here for AutoMigrate
	"email_server/utils"  // For HashPassword

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var DB *gorm.DB

func Init(dbPath string) { // dbPath could be from config
	var err error
	// DSN for SQLite is just the file path
	// Example: "gorm.db" or from config.AppConfig.Database.DSN if it's set to a file path
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}

	log.Println("✅ 数据库连接成功")

	// AutoMigrate models
	// Ensure all models defined in optimization_proposal.md are included here
	// For now, using a placeholder. This will be updated when models are defined.
	err = DB.AutoMigrate(
		&models.User{},
		&models.EmailAccount{},
		&models.Platform{},
		&models.PlatformRegistration{},
		&models.ServiceSubscription{},
		// Add other models here as they are defined
	)
	if err != nil {
		log.Fatal("❌ 数据库表自动迁移失败:", err)
	}
	log.Println("🎉 数据表自动迁移完成")

	// 创建默认管理员账户（如果不存在）
	createDefaultAdminUser()
}

func createDefaultAdminUser() {
	var adminUser models.User
	// 检查 "admin" 用户是否已存在
	err := DB.Where("username = ?", "admin").First(&adminUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户不存在，创建新用户
			hashedPassword, hashErr := utils.HashPassword("password")
			if hashErr != nil {
				log.Fatalf("❌ 创建默认管理员时密码哈希失败: %v", hashErr)
				return
			}
			defaultAdmin := models.User{
				Username: "admin",
				Email:    "admin@example.com", // 或者一个更合适的默认邮箱
				Password: hashedPassword,
				// Role: "admin", // 如果有 Role 字段
			}
			if createErr := DB.Create(&defaultAdmin).Error; createErr != nil {
				log.Fatalf("❌ 创建默认管理员账户失败: %v", createErr)
			} else {
				log.Println("🔑 默认管理员账户 'admin' 创建成功 (密码: password)")
			}
		} else {
			// 查询时发生其他错误
			log.Fatalf("❌ 查询默认管理员账户失败: %v", err)
		}
	} else {
		// 用户已存在
		log.Println("ℹ️ 默认管理员账户 'admin' 已存在.")
	}
}

// Close function might not be strictly necessary with GORM for typical app lifecycle,
// but can be kept if specific resource cleanup is needed.
// GORM's DB instance typically manages its connection pool.
// func Close() {
// 	if DB != nil {
// 		sqlDB, err := DB.DB()
// 		if err == nil {
// 			sqlDB.Close()
// 		}
// 	}
// }
