package database

import (
	"log"
	// "email_server/config" // Config will be handled differently or passed to Init
	"email_server/models" // Assuming models will be defined here for AutoMigrate
	"email_server/utils"  // For HashPassword

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

	// 应用自定义索引
	if err := applyCustomIndexes(DB); err != nil {
		// 根据 applyCustomIndexes 的具体实现，这里可能需要错误处理
		// 例如 log.Printf("警告: 应用自定义索引时发生错误: %v", err)
		// 或者如果错误是致命的: log.Fatalf("错误: 应用自定义索引失败: %v", err)
	}

	// 创建默认管理员账户（如果不存在）
	createDefaultAdminUser()
}

// applyCustomIndexes 应用数据库的自定义索引
// 注意：为了清晰和可测试性，将 DB 作为参数传递
func applyCustomIndexes(db *gorm.DB) error {
	// 约束 1: (user_id, platform_id, username) 的组合必须唯一，当 username 非空字符串时。
	sqlUsernameIndex := `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_user_platform_username_not_empty
	ON platform_registrations (user_id, platform_id, login_username)
	WHERE login_username IS NOT NULL AND login_username != '';`

	if err := db.Exec(sqlUsernameIndex).Error; err != nil {
		log.Printf("⚠️ 创建 username 部分唯一索引失败 (可能已存在或 SQLite 版本不支持): %v", err)
		// 返回错误，以便调用者可以决定如何处理
		// return fmt.Errorf("创建 username 唯一索引失败: %w", err)
	} else {
		log.Println("👍 Username 部分唯一索引创建/验证成功")
	}

	// 约束 2: (user_id, platform_id, email_account_id) 的组合必须唯一，当 email_account_id 非 NULL 且非0时。
	sqlEmailIndex := `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_user_platform_email_id_not_null_zero
	ON platform_registrations (user_id, platform_id, email_account_id)
	WHERE email_account_id IS NOT NULL AND email_account_id != 0;`

	if err := db.Exec(sqlEmailIndex).Error; err != nil {
		log.Printf("⚠️ 创建 email_account_id 部分唯一索引失败 (可能已存在或 SQLite 版本不支持): %v", err)
		// return fmt.Errorf("创建 email_account_id 唯一索引失败: %w", err)
	} else {
		log.Println("👍 EmailAccountID 部分唯一索引创建/验证成功")
	}

	// 约束 3: 平台名称唯一性约束，只对未删除的记录生效
	// 首先删除旧的唯一索引（如果存在）
	dropOldPlatformIndex := `DROP INDEX IF EXISTS idx_user_platform_name;`
	if err := db.Exec(dropOldPlatformIndex).Error; err != nil {
		log.Printf("⚠️ 删除旧的平台名称唯一索引失败: %v", err)
	} else {
		log.Println("👍 旧的平台名称唯一索引删除成功")
	}

	// 清理重复的平台数据（保留最新的记录）
	cleanupDuplicatePlatforms := `
	DELETE FROM platforms
	WHERE id NOT IN (
		SELECT MIN(id)
		FROM platforms
		WHERE deleted_at IS NULL
		GROUP BY user_id, name
	) AND deleted_at IS NULL;`

	if err := db.Exec(cleanupDuplicatePlatforms).Error; err != nil {
		log.Printf("⚠️ 清理重复平台数据失败: %v", err)
	} else {
		log.Println("👍 重复平台数据清理完成")
	}

	// 创建新的部分唯一索引，只对未删除的记录生效
	sqlPlatformIndex := `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_user_platform_name_not_deleted
	ON platforms (user_id, name)
	WHERE deleted_at IS NULL;`

	if err := db.Exec(sqlPlatformIndex).Error; err != nil {
		log.Printf("⚠️ 创建平台名称部分唯一索引失败 (可能已存在或 SQLite 版本不支持): %v", err)
		// return fmt.Errorf("创建平台名称唯一索引失败: %w", err)
	} else {
		log.Println("👍 平台名称部分唯一索引创建/验证成功")
	}

	// 约束 4: 邮箱账户唯一性约束，只对未删除的记录生效
	// 首先删除旧的唯一索引（如果存在）
	dropOldEmailIndex := `DROP INDEX IF EXISTS idx_user_email;`
	if err := db.Exec(dropOldEmailIndex).Error; err != nil {
		log.Printf("⚠️ 删除旧的邮箱账户唯一索引失败: %v", err)
	} else {
		log.Println("👍 旧的邮箱账户唯一索引删除成功")
	}

	// 创建新的部分唯一索引，只对未删除的记录生效
	sqlEmailAccountIndex := `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_user_email_not_deleted
	ON email_accounts (user_id, email_address)
	WHERE deleted_at IS NULL;`

	if err := db.Exec(sqlEmailAccountIndex).Error; err != nil {
		log.Printf("⚠️ 创建邮箱账户部分唯一索引失败 (可能已存在或 SQLite 版本不支持): %v", err)
		// return fmt.Errorf("创建邮箱账户唯一索引失败: %w", err)
	} else {
		log.Println("👍 邮箱账户部分唯一索引创建/验证成功")
	}

	return nil // 如果所有索引都成功或允许部分失败，则返回 nil
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
				Role:     models.RoleAdmin,    // 设置为管理员角色
				Status:   models.StatusActive, // 设置为激活状态
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
