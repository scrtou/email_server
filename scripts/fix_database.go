package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 从环境变量获取数据库路径，默认为 /data/database.db
	dbPath := os.Getenv("SQLITE_FILE")
	if dbPath == "" {
		dbPath = "/data/database.db"
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}

	log.Println("✅ 数据库连接成功")

	// 删除旧的唯一索引
	oldIndexes := []string{
		"idx_unique_user_platform_username_not_empty",
		"idx_unique_user_platform_email_id_not_null_zero",
		"idx_user_platform_name",
		"idx_unique_user_platform_name_not_deleted",
		"idx_user_email",
		"idx_unique_user_email_not_deleted",
	}

	for _, indexName := range oldIndexes {
		dropSQL := fmt.Sprintf("DROP INDEX IF EXISTS %s;", indexName)
		if err := db.Exec(dropSQL).Error; err != nil {
			log.Printf("⚠️ 删除旧索引 %s 失败: %v", indexName, err)
		} else {
			log.Printf("👍 旧索引 %s 删除成功", indexName)
		}
	}

	log.Println("🎉 数据库索引修复完成")
	log.Println("现在软删除记录不会阻止创建新记录了")
}
