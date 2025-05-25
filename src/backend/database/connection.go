package database

import (
    "database/sql"
    "log"

    _ "github.com/go-sql-driver/mysql"
    "email_server/config"
)

var DB *sql.DB

func Init() {
    var err error
    DB, err = sql.Open("mysql", config.AppConfig.Database.DSN)
    if err != nil {
        log.Fatal("数据库连接失败:", err)
    }

    if err = DB.Ping(); err != nil {
        log.Fatal("数据库连接测试失败:", err)
    }

    log.Println("数据库连接成功")
    CreateTables()
}

func Close() {
    if DB != nil {
        DB.Close()
    }
}

func CreateTables() {
    // 创建用户表 - 添加这个
    createUsersTable()
    
    // 原有的表
    createEmailsTable()
    createServicesTable()
    createEmailServicesTable()
    
    log.Println("🎉 数据表初始化完成")
}

// 新增：创建用户表
func createUsersTable() {
    userTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        username VARCHAR(50) UNIQUE NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        real_name VARCHAR(100) NOT NULL,
        phone VARCHAR(20) DEFAULT '' COMMENT '手机号',
        role ENUM('admin', 'user') DEFAULT 'user',
        status TINYINT DEFAULT 1 COMMENT '1:正常 0:禁用',
        last_login TIMESTAMP NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        INDEX idx_username (username),
        INDEX idx_email (email),
        INDEX idx_status (status)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';`
    
    if _, err := DB.Exec(userTableSQL); err != nil {
        log.Printf("❌ 创建用户表失败: %v", err)
    } else {
        log.Printf("✅ 用户表准备完成")
        createDefaultAdmin()
    }
}


// 创建默认管理员账户
func createDefaultAdmin() {
    var count int
    err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
    if err != nil || count > 0 {
        return
    }
    
    // 导入密码工具
    hashedPassword := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // 对应密码: password
    
    query := `
        INSERT INTO users (username, email, password, real_name, role) 
        VALUES ('admin', 'admin@example.com', ?, '系统管理员', 'admin')
    `
    _, err = DB.Exec(query, hashedPassword)
    if err != nil {
        log.Printf("❌ 创建默认管理员失败: %v", err)
    } else {
        log.Printf("✅ 默认管理员账户创建成功 (用户名: admin, 密码: password)")
    }
}

// 修改邮箱表，添加用户关联
func createEmailsTable() {
    emailTableSQL := `
    CREATE TABLE IF NOT EXISTS emails (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        user_id BIGINT NOT NULL COMMENT '所属用户ID',
        email VARCHAR(255) NOT NULL,
        password VARCHAR(255),
        display_name VARCHAR(100),
        provider VARCHAR(50),
        phone VARCHAR(20),
        backup_email VARCHAR(255),
        security_question VARCHAR(500),
        security_answer VARCHAR(255),
        notes TEXT,
        status TINYINT DEFAULT 1,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
        UNIQUE KEY unique_user_email (user_id, email),
        INDEX idx_user_id (user_id),
        INDEX idx_email (email),
        INDEX idx_status (status)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮箱表';`

    if _, err := DB.Exec(emailTableSQL); err != nil {
        log.Printf("❌ 创建邮箱表失败: %v", err)
    } else {
        log.Printf("✅ 邮箱表准备完成")
    }
}

func createServicesTable() {
    serviceTableSQL := `
    CREATE TABLE IF NOT EXISTS services (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(100) NOT NULL,
        website VARCHAR(255),
        category VARCHAR(50),
        description TEXT,
        logo_url VARCHAR(255),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        INDEX idx_name (name),
        INDEX idx_category (category)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务表';`

    if _, err := DB.Exec(serviceTableSQL); err != nil {
        log.Printf("❌ 创建服务表失败: %v", err)
    } else {
        log.Printf("✅ 服务表准备完成")
    }
}

func createEmailServicesTable() {
    emailServiceTableSQL := `
    CREATE TABLE IF NOT EXISTS email_services (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        email_id BIGINT NOT NULL,
        service_id BIGINT NOT NULL,
        username VARCHAR(255),
        password VARCHAR(255),
        phone VARCHAR(20),
        registration_date DATE,
        subscription_type VARCHAR(50),
        subscription_expires DATE,
        notes TEXT,
        status TINYINT DEFAULT 1,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE,
        FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
        UNIQUE KEY unique_email_service (email_id, service_id),
        INDEX idx_email_id (email_id),
        INDEX idx_service_id (service_id),
        INDEX idx_status (status)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮箱服务关联表';`

    if _, err := DB.Exec(emailServiceTableSQL); err != nil {
        log.Printf("❌ 创建邮箱服务关联表失败: %v", err)
    } else {
        log.Printf("✅ 邮箱服务关联表准备完成")
    }
}
