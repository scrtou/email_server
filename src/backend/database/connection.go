package database

import (
	"log"

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
	)
	if err != nil {
		log.Fatal("❌ 数据库表自动迁移失败:", err)
	}
	log.Println("🎉 数据表自动迁移完成")

	// 创建默认管理员账户
	createDefaultAdminUser()
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
