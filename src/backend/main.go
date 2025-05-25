package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"os"
	"path/filepath"
	"strings" 
	"time"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// 数据模型保持不变...
type Email struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Password     string    `json:"password,omitempty" db:"password"`
	DisplayName  string    `json:"display_name" db:"display_name"`
	Provider     string    `json:"provider" db:"provider"`
	Phone        string    `json:"phone" db:"phone"`
	BackupEmail  string    `json:"backup_email" db:"backup_email"`
	SecurityQ    string    `json:"security_question" db:"security_question"`
	SecurityA    string    `json:"security_answer,omitempty" db:"security_answer"`
	Notes        string    `json:"notes" db:"notes"`
	Status       int       `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	ServiceCount int       `json:"service_count,omitempty"`
}

type Service struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Website     string    `json:"website" db:"website"`
	Category    string    `json:"category" db:"category"`
	Description string    `json:"description" db:"description"`
	LogoURL     string    `json:"logo_url" db:"logo_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	EmailCount  int       `json:"email_count,omitempty"`
}

type EmailService struct {
	ID                  int64      `json:"id" db:"id"`
	EmailID             int64      `json:"email_id" db:"email_id"`
	ServiceID           int64      `json:"service_id" db:"service_id"`
	Username            string     `json:"username" db:"username"`
	Password            string     `json:"password,omitempty" db:"password"`
	Phone               string     `json:"phone" db:"phone"`
	RegistrationDate    *time.Time `json:"registration_date" db:"registration_date"`
	SubscriptionType    string     `json:"subscription_type" db:"subscription_type"`
	SubscriptionExpires *time.Time `json:"subscription_expires" db:"subscription_expires"`
	Notes               string     `json:"notes" db:"notes"`
	Status              int        `json:"status" db:"status"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	Email       *Email   `json:"email,omitempty"`
	Service     *Service `json:"service,omitempty"`
	ServiceName string   `json:"service_name,omitempty"`
	EmailAddr   string   `json:"email_addr,omitempty"`
}

type DashboardData struct {
	EmailCount         int            `json:"email_count"`
	ServiceCount       int            `json:"service_count"`
	RelationCount      int            `json:"relation_count"`
	EmailsByProvider   map[string]int `json:"emails_by_provider"`
	ServicesByCategory map[string]int `json:"services_by_category"`
	RecentEmails       []*Email       `json:"recent_emails"`
	RecentServices     []*Service     `json:"recent_services"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 全局数据库连接
var db *sql.DB

// 初始化数据库
func initDB() {
	var err error
	dsn := "avnadmin:AVNS_icoPVWCDqQgoAM4nCH1@tcp(mysql-yxmysql.c.aivencloud.com:19894)/email-server?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	log.Println("数据库连接成功")
	createTables()
}

// 创建数据表
func createTables() {
	emailTableSQL := `
	CREATE TABLE IF NOT EXISTS emails (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		email VARCHAR(255) UNIQUE NOT NULL,
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
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`

	serviceTableSQL := `
	CREATE TABLE IF NOT EXISTS services (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(100) NOT NULL,
		website VARCHAR(255),
		category VARCHAR(50),
		description TEXT,
		logo_url VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`

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
		UNIQUE KEY unique_email_service (email_id, service_id)
	);`

	tables := map[string]string{
		"emails":         emailTableSQL,
		"services":       serviceTableSQL,
		"email_services": emailServiceTableSQL,
	}

	for tableName, tableSQL := range tables {
		log.Printf("正在检查并创建表: %s", tableName)
		if _, err := db.Exec(tableSQL); err != nil {
			log.Printf("❌ 创建表 %s 失败: %v", tableName, err)
		} else {
			log.Printf("✅ 表 %s 准备完成", tableName)
		}
	}
	
	log.Println("🎉 数据表初始化完成")
}

// 响应工具函数
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

func sendError(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// 保持原有的所有函数不变...
func getDashboard(c *gin.Context) {
	dashboard := &DashboardData{
		EmailsByProvider:   make(map[string]int),
		ServicesByCategory: make(map[string]int),
	}

	err := db.QueryRow("SELECT COUNT(*) FROM emails WHERE status = 1").Scan(&dashboard.EmailCount)
	if err != nil {
		sendError(c, 500, "获取邮箱数量失败")
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM services").Scan(&dashboard.ServiceCount)
	if err != nil {
		sendError(c, 500, "获取服务数量失败")
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM email_services WHERE status = 1").Scan(&dashboard.RelationCount)
	if err != nil {
		sendError(c, 500, "获取关联数量失败")
		return
	}

	rows, err := db.Query("SELECT provider, COUNT(*) FROM emails WHERE status = 1 GROUP BY provider")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var provider string
			var count int
			if err := rows.Scan(&provider, &count); err == nil {
				dashboard.EmailsByProvider[provider] = count
			}
		}
	}

	rows, err = db.Query("SELECT category, COUNT(*) FROM services GROUP BY category")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var category string
			var count int
			if err := rows.Scan(&category, &count); err == nil {
				dashboard.ServicesByCategory[category] = count
			}
		}
	}

	dashboard.RecentEmails = getRecentEmails()
	dashboard.RecentServices = getRecentServices()

	success(c, dashboard)
}

// ✅ 新增：获取所有邮箱服务关联的函数
func getAllEmailServices(c *gin.Context) {
	// 支持分页参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")
	
	// 支持过滤参数
	emailID := c.Query("email_id")
	serviceID := c.Query("service_id")
	status := c.DefaultQuery("status", "1")

	// 构建查询
	query := `
		SELECT 
			es.id, es.email_id, es.service_id, es.username, es.phone,
			es.registration_date, es.subscription_type, es.subscription_expires,
			es.notes, es.status, es.created_at, es.updated_at,
			e.email as email_addr, e.display_name as email_display_name,
			s.name as service_name, s.website as service_website
		FROM email_services es
		LEFT JOIN emails e ON es.email_id = e.id
		LEFT JOIN services s ON es.service_id = s.id
		WHERE es.status = ?
	`
	args := []interface{}{status}

	// 添加过滤条件
	if emailID != "" {
		query += " AND es.email_id = ?"
		args = append(args, emailID)
	}
	if serviceID != "" {
		query += " AND es.service_id = ?"
		args = append(args, serviceID)
	}

	// 添加排序
	query += " ORDER BY es.created_at DESC"
	
	// 如果需要分页
	if page != "all" {
		pageNum, _ := strconv.Atoi(page)
		pageSizeNum, _ := strconv.Atoi(pageSize)
		offset := (pageNum - 1) * pageSizeNum
		query += " LIMIT ? OFFSET ?"
		args = append(args, pageSizeNum, offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("查询邮箱服务关联失败: %v", err)
		sendError(c, 500, "查询关联数据失败")
		return
	}
	defer rows.Close()

	var emailServices []*EmailService
	for rows.Next() {
		es := &EmailService{}
		var emailDisplayName, serviceWebsite sql.NullString
		
		err := rows.Scan(
			&es.ID, &es.EmailID, &es.ServiceID, &es.Username, &es.Phone,
			&es.RegistrationDate, &es.SubscriptionType, &es.SubscriptionExpires,
			&es.Notes, &es.Status, &es.CreatedAt, &es.UpdatedAt,
			&es.EmailAddr, &emailDisplayName, &es.ServiceName, &serviceWebsite,
		)
		if err != nil {
			log.Printf("扫描关联数据失败: %v", err)
			continue
		}
		emailServices = append(emailServices, es)
	}

	// 如果需要返回总数（用于分页）
	if page != "all" {
		countQuery := `
			SELECT COUNT(*) 
			FROM email_services es
			WHERE es.status = ?
		`
		countArgs := []interface{}{status}
		
		if emailID != "" {
			countQuery += " AND es.email_id = ?"
			countArgs = append(countArgs, emailID)
		}
		if serviceID != "" {
			countQuery += " AND es.service_id = ?"
			countArgs = append(countArgs, serviceID)
		}

		var total int
		err = db.QueryRow(countQuery, countArgs...).Scan(&total)
		if err != nil {
			log.Printf("查询总数失败: %v", err)
		}

		// 返回分页数据
		result := map[string]interface{}{
			"data":       emailServices,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		}
		success(c, result)
		return
	}

	success(c, emailServices)
}

// 保持所有原有函数不变...
func getRecentEmails() []*Email {
	query := `
		SELECT e.id, e.email, e.display_name, e.provider, e.phone, e.created_at, COUNT(es.id) as service_count
		FROM emails e
		LEFT JOIN email_services es ON e.id = es.email_id AND es.status = 1
		WHERE e.status = 1
		GROUP BY e.id
		ORDER BY e.created_at DESC
		LIMIT 5
	`
	rows, err := db.Query(query)
	if err != nil {
		return []*Email{}
	}
	defer rows.Close()

	var emails []*Email
	for rows.Next() {
		email := &Email{}
		err := rows.Scan(&email.ID, &email.Email, &email.DisplayName, &email.Provider,
			&email.Phone, &email.CreatedAt, &email.ServiceCount)
		if err == nil {
			emails = append(emails, email)
		}
	}
	return emails
}

func getRecentServices() []*Service {
	query := `
		SELECT s.id, s.name, s.website, s.category, s.description, s.created_at, COUNT(es.id) as email_count
		FROM services s
		LEFT JOIN email_services es ON s.id = es.service_id AND es.status = 1
		GROUP BY s.id
		ORDER BY s.created_at DESC
		LIMIT 5
	`
	rows, err := db.Query(query)
	if err != nil {
		return []*Service{}
	}
	defer rows.Close()

	var services []*Service
	for rows.Next() {
		service := &Service{}
		err := rows.Scan(&service.ID, &service.Name, &service.Website, &service.Category,
			&service.Description, &service.CreatedAt, &service.EmailCount)
		if err == nil {
			services = append(services, service)
		}
	}
	return services
}

func getEmails(c *gin.Context) {
	email := c.Query("email")
	provider := c.Query("provider")

	query := `
		SELECT e.id, e.email, e.display_name, e.provider, e.phone, e.backup_email, 
			   e.notes, e.status, e.created_at, e.updated_at, COUNT(es.id) as service_count
		FROM emails e
		LEFT JOIN email_services es ON e.id = es.email_id AND es.status = 1
		WHERE e.status = 1
	`
	args := []interface{}{}

	if email != "" {
		query += " AND e.email LIKE ?"
		args = append(args, "%"+email+"%")
	}
	if provider != "" {
		query += " AND e.provider = ?"
		args = append(args, provider)
	}

	query += " GROUP BY e.id ORDER BY e.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		sendError(c, 500, "查询邮箱失败")
		return
	}
	defer rows.Close()

	var emails []*Email
	for rows.Next() {
		email := &Email{}
		err := rows.Scan(&email.ID, &email.Email, &email.DisplayName, &email.Provider,
			&email.Phone, &email.BackupEmail, &email.Notes, &email.Status,
			&email.CreatedAt, &email.UpdatedAt, &email.ServiceCount)
		if err == nil {
			emails = append(emails, email)
		}
	}

	success(c, emails)
}

func createEmail(c *gin.Context) {
	var email Email
	if err := c.ShouldBindJSON(&email); err != nil {
		sendError(c, 400, "参数错误")
		return
	}

	query := `
		INSERT INTO emails (email, password, display_name, provider, phone, backup_email, 
			security_question, security_answer, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query, email.Email, email.Password, email.DisplayName,
		email.Provider, email.Phone, email.BackupEmail, email.SecurityQ,
		email.SecurityA, email.Notes)
	if err != nil {
		sendError(c, 500, "创建邮箱失败")
		return
	}

	id, _ := result.LastInsertId()
	email.ID = id
	success(c, email)
}

func updateEmail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	var email Email
	if err := c.ShouldBindJSON(&email); err != nil {
		sendError(c, 400, "参数错误")
		return
	}

	query := `
		UPDATE emails SET email=?, password=?, display_name=?, provider=?, phone=?, 
			backup_email=?, security_question=?, security_answer=?, notes=?, updated_at=NOW()
		WHERE id=?
	`
	_, err = db.Exec(query, email.Email, email.Password, email.DisplayName,
		email.Provider, email.Phone, email.BackupEmail, email.SecurityQ,
		email.SecurityA, email.Notes, id)
	if err != nil {
		sendError(c, 500, "更新邮箱失败")
		return
	}

	success(c, "更新成功")
}

func deleteEmail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	_, err = db.Exec("UPDATE emails SET status=0 WHERE id=?", id)
	if err != nil {
		sendError(c, 500, "删除邮箱失败")
		return
	}

	success(c, "删除成功")
}

func getEmailByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	var email Email
	query := `
		SELECT id, email, display_name, provider, phone, backup_email, 
			   security_question, notes, status, created_at, updated_at
		FROM emails WHERE id=? AND status=1
	`
	err = db.QueryRow(query, id).Scan(&email.ID, &email.Email, &email.DisplayName,
		&email.Provider, &email.Phone, &email.BackupEmail, &email.SecurityQ,
		&email.Notes, &email.Status, &email.CreatedAt, &email.UpdatedAt)
	if err != nil {
		sendError(c, 404, "邮箱不存在")
		return
	}

	success(c, email)
}

// ✅ 重命名为 getEmailServices（获取特定邮箱的服务）
func getEmailServices(c *gin.Context) {
	emailID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	query := `
		SELECT es.id, es.email_id, es.service_id, es.username, es.phone, 
			   es.registration_date, es.subscription_type, es.subscription_expires,
			   es.notes, es.status, es.created_at, es.updated_at, s.name as service_name
		FROM email_services es
		JOIN services s ON es.service_id = s.id
		WHERE es.email_id = ? AND es.status = 1
		ORDER BY es.created_at DESC
	`
	rows, err := db.Query(query, emailID)
	if err != nil {
		sendError(c, 500, "查询邮箱服务失败")
		return
	}
	defer rows.Close()

	var emailServices []*EmailService
	for rows.Next() {
		es := &EmailService{}
		err := rows.Scan(&es.ID, &es.EmailID, &es.ServiceID, &es.Username, &es.Phone,
			&es.RegistrationDate, &es.SubscriptionType, &es.SubscriptionExpires,
			&es.Notes, &es.Status, &es.CreatedAt, &es.UpdatedAt, &es.ServiceName)
		if err == nil {
			emailServices = append(emailServices, es)
		}
	}

	success(c, emailServices)
}

func getServices(c *gin.Context) {
	name := c.Query("name")
	category := c.Query("category")

	query := `
		SELECT s.id, s.name, s.website, s.category, s.description, s.logo_url,
			   s.created_at, s.updated_at, COUNT(es.id) as email_count
		FROM services s
		LEFT JOIN email_services es ON s.id = es.service_id AND es.status = 1
		WHERE 1=1
	`
	args := []interface{}{}

	if name != "" {
		query += " AND s.name LIKE ?"
		args = append(args, "%"+name+"%")
	}
	if category != "" {
		query += " AND s.category = ?"
		args = append(args, category)
	}

	query += " GROUP BY s.id ORDER BY s.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		sendError(c, 500, "查询服务失败")
		return
	}
	defer rows.Close()

	var services []*Service
	for rows.Next() {
		service := &Service{}
		err := rows.Scan(&service.ID, &service.Name, &service.Website, &service.Category,
			&service.Description, &service.LogoURL, &service.CreatedAt, &service.UpdatedAt,
			&service.EmailCount)
		if err == nil {
			services = append(services, service)
		}
	}

	success(c, services)
}

func createService(c *gin.Context) {
	var service Service
	if err := c.ShouldBindJSON(&service); err != nil {
		sendError(c, 400, "参数错误")
		return
	}

	query := `
		INSERT INTO services (name, website, category, description, logo_url)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query, service.Name, service.Website, service.Category,
		service.Description, service.LogoURL)
	if err != nil {
		sendError(c, 500, "创建服务失败")
		return
	}

	id, _ := result.LastInsertId()
	service.ID = id
	success(c, service)
}

func updateService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	var service Service
	if err := c.ShouldBindJSON(&service); err != nil {
		sendError(c, 400, "参数错误")
		return
	}

	query := `
		UPDATE services SET name=?, website=?, category=?, description=?, 
			logo_url=?, updated_at=NOW()
		WHERE id=?
	`
	_, err = db.Exec(query, service.Name, service.Website, service.Category,
		service.Description, service.LogoURL, id)
	if err != nil {
		sendError(c, 500, "更新服务失败")
		return
	}

	success(c, "更新成功")
}

func deleteService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	_, err = db.Exec("DELETE FROM services WHERE id=?", id)
	if err != nil {
		sendError(c, 500, "删除服务失败")
		return
	}

	success(c, "删除成功")
}

func getServiceByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	var service Service
	err = db.QueryRow("SELECT id, name, website, category, description, logo_url, created_at, updated_at FROM services WHERE id=?", id).
		Scan(&service.ID, &service.Name, &service.Website, &service.Category,
			&service.Description, &service.LogoURL, &service.CreatedAt, &service.UpdatedAt)
	if err != nil {
		sendError(c, 404, "服务不存在")
		return
	}

	success(c, service)
}

func getServiceEmails(c *gin.Context) {
	serviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	query := `
		SELECT es.id, es.email_id, es.service_id, es.username, es.phone,
			   es.registration_date, es.subscription_type, es.subscription_expires,
			   es.notes, es.status, es.created_at, es.updated_at, e.email
		FROM email_services es
		JOIN emails e ON es.email_id = e.id
		WHERE es.service_id = ? AND es.status = 1 AND e.status = 1
		ORDER BY es.created_at DESC
	`
	rows, err := db.Query(query, serviceID)
	if err != nil {
		sendError(c, 500, "查询服务邮箱失败")
		return
	}
	defer rows.Close()

	var serviceEmails []*EmailService
	for rows.Next() {
		se := &EmailService{}
		err := rows.Scan(&se.ID, &se.EmailID, &se.ServiceID, &se.Username, &se.Phone,
			&se.RegistrationDate, &se.SubscriptionType, &se.SubscriptionExpires,
			&se.Notes, &se.Status, &se.CreatedAt, &se.UpdatedAt, &se.EmailAddr)
		if err == nil {
			serviceEmails = append(serviceEmails, se)
		}
	}

	success(c, serviceEmails)
}

func createEmailService(c *gin.Context) {
	var es EmailService
	if err := c.ShouldBindJSON(&es); err != nil {
		sendError(c, 400, "参数错误")
		return
	}

	query := `
		INSERT INTO email_services (email_id, service_id, username, password, phone,
			registration_date, subscription_type, subscription_expires, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query, es.EmailID, es.ServiceID, es.Username, es.Password,
		es.Phone, es.RegistrationDate, es.SubscriptionType, es.SubscriptionExpires, es.Notes)
	if err != nil {
		sendError(c, 500, "创建关联失败")
		return
	}

	id, _ := result.LastInsertId()
	es.ID = id
	success(c, es)
}

func updateEmailService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	var es EmailService
	if err := c.ShouldBindJSON(&es); err != nil {
		sendError(c, 400, "参数错误")
		return
	}

	query := `
		UPDATE email_services SET username=?, password=?, phone=?, registration_date=?,
			subscription_type=?, subscription_expires=?, notes=?, updated_at=NOW()
		WHERE id=?
	`
	_, err = db.Exec(query, es.Username, es.Password, es.Phone, es.RegistrationDate,
		es.SubscriptionType, es.SubscriptionExpires, es.Notes, id)
	if err != nil {
		sendError(c, 500, "更新关联失败")
		return
	}

	success(c, "更新成功")
}

func deleteEmailService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		sendError(c, 400, "ID参数错误")
		return
	}

	_, err = db.Exec("UPDATE email_services SET status=0 WHERE id=?", id)
	if err != nil {
		sendError(c, 500, "删除关联失败")
		return
	}

	success(c, "删除成功")
}

// 静态文件服务中间件
func serveStaticFiles(r *gin.Engine) {
	distPath := "../frontend/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		log.Println("⚠️  前端构建文件不存在，跳过静态文件服务")
		log.Println("请运行: cd frontend && npm run build")
		return
	}

	log.Printf("✅ 找到前端构建文件: %s", distPath)

	r.Static("/static", filepath.Join(distPath, "static"))
	r.Static("/js", filepath.Join(distPath, "js"))
	r.Static("/css", filepath.Join(distPath, "css"))
	r.StaticFile("/favicon.ico", filepath.Join(distPath, "favicon.ico"))
	
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(404, gin.H{
				"code":    404,
				"message": "API endpoint not found",
				"path":    path,
			})
			return
		}
		
		if strings.Contains(path, ".") {
			c.Status(404)
			return
		}
		
		indexPath := filepath.Join(distPath, "index.html")
		log.Printf("🏠 返回前端页面: %s -> %s", path, indexPath)
		c.File(indexPath)
	})
}

// ✅ 修复：setupRouter函数
func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // AllowAllOrigins为true时必须为false
	}))

	api := r.Group("/api/v1")
	{
		api.GET("/dashboard", getDashboard)

		emails := api.Group("/emails")
		{
			emails.GET("", getEmails)
			emails.POST("", createEmail)
			emails.GET("/:id", getEmailByID)
			emails.PUT("/:id", updateEmail)
			emails.DELETE("/:id", deleteEmail)
			emails.GET("/:id/services", getEmailServices)
		}

		services := api.Group("/services")
		{
			services.GET("", getServices)
			services.POST("", createService)
			services.GET("/:id", getServiceByID)
			services.PUT("/:id", updateService)
			services.DELETE("/:id", deleteService)
			services.GET("/:id/emails", getServiceEmails)
		}

		// ✅ 修复：邮箱服务关联管理
		emailServices := api.Group("/email-services")
		{
			emailServices.GET("", getAllEmailServices)          // ✅ 修改为getAllEmailServices
			emailServices.POST("", createEmailService)
			emailServices.PUT("/:id", updateEmailService)
			emailServices.DELETE("/:id", deleteEmailService)
		}
	} // ✅ 移除多余的}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now()})
	})

	serveStaticFiles(r)

	return r
}

func main() {
	initDB()
	defer db.Close()

	r := setupRouter()

	fmt.Println("服务器启动在 http://localhost:5555")
	log.Fatal(r.Run(":5555"))
}
