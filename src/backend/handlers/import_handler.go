package handlers

import (
	"email_server/database" // Import database package
	"email_server/importer"
	"email_server/models" // Import models package
	"email_server/utils"
	"encoding/json" // Import json package
	"errors"        // Added for errors.Is
	"fmt"           // Import fmt package
	"log"
	"net/http"
	"net/mail" // Added for email validation
	"strconv"
	"strings" // Import strings package

	"github.com/gin-gonic/gin"
	"gorm.io/gorm" // Import gorm package
)

// ImportBitwardenCSVHandler 处理 Bitwarden CSV 文件导入请求
func ImportBitwardenCSVHandler(c *gin.Context) {
	// 检查 Content-Type 是否为 multipart/form-data
	if c.ContentType() != "multipart/form-data" {
		log.Printf("ImportBitwardenCSVHandler: Invalid Content-Type: %s", c.ContentType())
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求必须是 multipart/form-data 类型")
		return
	}

	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Printf("ImportBitwardenCSVHandler: Error getting form file 'file': %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "无法获取上传的文件: "+err.Error())
		return
	}

	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("ImportBitwardenCSVHandler: Error opening uploaded file: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "无法打开上传的文件: "+err.Error())
		return
	}
	defer file.Close() // 确保文件被关闭

	// 获取 importPasswords 参数 (默认为 false)
	importPasswordsStr := c.PostForm("importPasswords")
	importPasswords, _ := strconv.ParseBool(importPasswordsStr) // 忽略错误，默认为 false

	log.Printf("ImportBitwardenCSVHandler: Received file '%s', importPasswords=%t", fileHeader.Filename, importPasswords)

	// 调用解析器
	items, err := importer.ParseBitwardenCSV(file, importPasswords)
	if err != nil {
		log.Printf("ImportBitwardenCSVHandler: Error parsing Bitwarden CSV: %v", err)
		// 根据错误类型判断是客户端错误还是服务端错误
		// 简单的处理：假设大部分解析错误是由于文件格式问题 (客户端错误)
		utils.SendErrorResponse(c, http.StatusBadRequest, "解析 CSV 文件失败: "+err.Error())
		return
	}

	log.Printf("ImportBitwardenCSVHandler: Successfully parsed %d items from '%s'. Now attempting to save to database.", len(items), fileHeader.Filename)

	// --- 开始数据库保存逻辑 ---
	db := database.DB                       // 直接使用包级变量 DB
	userIDValue, exists := c.Get("user_id") // 使用正确的键名 "user_id"
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "无法获取用户信息 (context missing)")
		return
	}

	var userID uint
	switch v := userIDValue.(type) {
	case float64: // JWT claims often parse numbers as float64
		userID = uint(v)
	case int64:
		userID = uint(v)
	case uint:
		userID = v
	default:
		log.Printf("Unexpected type for user_id in context: %T", userIDValue)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID格式错误 (unexpected type)")
		return
	}

	if userID == 0 { // Basic check after conversion
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID无效 (zero value)")
		return
	}

	savedCount := 0
	errorCount := 0
	var errorMessages []string

	for i, item := range items {
		rowIndex := i + 2 // CSV data starts from the second row, and i is 0-indexed.

		// --- Input Validation from CSV item ---
		platformName := strings.TrimSpace(item.ItemName)
		loginIdentifier := strings.TrimSpace(item.Username) // This can be a username or an email

		if platformName == "" {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第 %d 行: 平台名称为空 (原始登录名: %s)", rowIndex, loginIdentifier))
			log.Printf("Import: Row %d skipped, platform name empty (LoginIdentifier: %s)", rowIndex, loginIdentifier)
			continue
		}
		if loginIdentifier == "" {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第 %d 行 (平台 %s): 登录标识符 (用户名/邮箱) 为空", rowIndex, platformName))
			log.Printf("Import: Row %d skipped, login identifier empty (Platform: %s)", rowIndex, platformName)
			continue
		}

		var errLoop error // Loop-scoped error variable

		// 1. Find or Create Platform (associated with the current user)
		var platform models.Platform
		errLoop = db.Where("name = ? AND user_id = ?", platformName, userID).First(&platform).Error
		if errLoop != nil {
			if errors.Is(errLoop, gorm.ErrRecordNotFound) {
				platform = models.Platform{
					UserID:     userID,
					Name:       platformName,
					WebsiteURL: item.URL, // Assign the URL from the imported item
				}
				if createErr := db.Create(&platform).Error; createErr != nil {
					errorCount++
					errorMessages = append(errorMessages, fmt.Sprintf("第 %d 行: 创建平台 '%s' 失败: %v", rowIndex, platformName, createErr))
					log.Printf("Import: Row %d error creating platform '%s': %v", rowIndex, platformName, createErr)
					continue
				}
				log.Printf("Import: Row %d created new platform '%s' (ID: %d) for user %d", rowIndex, platform.Name, platform.ID, userID)
			} else {
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("第 %d 行: 查询平台 '%s' 失败: %v", rowIndex, platformName, errLoop))
				log.Printf("Import: Row %d error querying platform '%s': %v", rowIndex, platformName, errLoop)
				continue
			}
		} else {
			// 找到了平台记录，直接使用
			log.Printf("Import: Row %d found existing platform '%s' (ID: %d) for user %d", rowIndex, platform.Name, platform.ID, userID)
		}
		// Platform is now valid, active, and associated with the user.

		// 2. Determine EmailAccount and LoginUsername for Registration
		var emailAccount models.EmailAccount
		var currentEmailAccountIDPtr *uint
		loginUsernameForRegistration := loginIdentifier // Default: treat loginIdentifier as username

		_, emailFormatErr := mail.ParseAddress(loginIdentifier)
		isEmail := emailFormatErr == nil

		if isEmail {
			emailAddress := loginIdentifier
			// loginUsernameForRegistration remains loginIdentifier as per user's latest instruction.
			// The original loginIdentifier will be stored as PlatformRegistration.LoginUsername.
			// If it's an email, it will ALSO be used to link/create an EmailAccount.

			errLoop = db.Where("email_address = ? AND user_id = ?", emailAddress, userID).First(&emailAccount).Error
			if errLoop != nil {
				if errors.Is(errLoop, gorm.ErrRecordNotFound) {
					emailAccount = models.EmailAccount{
						UserID:       userID,
						EmailAddress: emailAddress,
						Provider:     utils.ExtractProviderFromEmail(emailAddress),
					}
					if createErr := db.Create(&emailAccount).Error; createErr != nil {
						errorCount++
						errorMessages = append(errorMessages, fmt.Sprintf("第 %d 行: 创建邮箱账户 '%s' 失败: %v", rowIndex, emailAddress, createErr))
						log.Printf("Import: Row %d error creating email account '%s': %v", rowIndex, emailAddress, createErr)
						continue
					}
					log.Printf("Import: Row %d created new email account '%s' (ID: %d) for user %d", rowIndex, emailAccount.EmailAddress, emailAccount.ID, userID)
				} else {
					errorCount++
					errorMessages = append(errorMessages, fmt.Sprintf("第 %d 行: 查询邮箱账户 '%s' 失败: %v", rowIndex, emailAddress, errLoop))
					log.Printf("Import: Row %d error querying email account '%s': %v", rowIndex, emailAddress, errLoop)
					continue
				}
			} else {
				// 找到了邮箱账户记录，直接使用
				log.Printf("Import: Row %d found existing email account '%s' (ID: %d)", rowIndex, emailAccount.EmailAddress, emailAccount.ID)
			}
			// EmailAccount is now valid and active.
			if emailAccount.ID > 0 { // Make sure we have a valid ID
				tmpID := emailAccount.ID
				currentEmailAccountIDPtr = &tmpID
			}
		}

		// 3. Encrypt Password (if importPasswords is true)
		var encryptedPassword string
		if importPasswords && item.Password != "" {
			encryptedPassword, errLoop = utils.EncryptPassword(item.Password)
			if errLoop != nil {
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("第 %d 行 (平台 %s, 登录名 %s): 密码加密失败: %v. 密码未导入。", rowIndex, platform.Name, loginIdentifier, errLoop))
				log.Printf("Import: Row %d error encrypting password for login '%s', platform '%s': %v. Password not imported.", rowIndex, loginIdentifier, platform.Name, errLoop)
				encryptedPassword = "" // Ensure password is not set if encryption failed
			}
		}

		// 4. Conflict Check for PlatformRegistration
		// Check if a registration already exists for this user, platform, and combination of login username/email.
		var existingRegistration models.PlatformRegistration
		conflictQuery := db.Model(&models.PlatformRegistration{}).
			Where("user_id = ? AND platform_id = ?", userID, platform.ID)

		conflictMsg := ""
		if loginUsernameForRegistration != "" && currentEmailAccountIDPtr != nil {
			// This case implies loginIdentifier was an email, and we also want to store it as a username (unusual for this logic path)
			// For now, this path means loginIdentifier was NOT an email, but we somehow have an email_id. This should be rare.
			// Let's assume if currentEmailAccountIDPtr is set, loginUsernameForRegistration should be ""
			// The logic above ensures loginUsernameForRegistration is "" if isEmail is true.
			// So this specific branch of conflict check might be simplified.
			// If isEmail = true: loginUsernameForRegistration = "", currentEmailAccountIDPtr is set.
			// If isEmail = false: loginUsernameForRegistration = loginIdentifier, currentEmailAccountIDPtr is nil.

			conflictQuery = conflictQuery.Where("login_username = ? AND email_account_id = ?", loginUsernameForRegistration, *currentEmailAccountIDPtr)
			conflictMsg = fmt.Sprintf("第 %d 行: 用户名 '%s' 和邮箱ID %d 的组合已在此平台注册。", rowIndex, loginUsernameForRegistration, *currentEmailAccountIDPtr)
		} else if loginUsernameForRegistration != "" { // loginIdentifier was not an email
			conflictQuery = conflictQuery.Where("login_username = ? AND (email_account_id IS NULL OR email_account_id = 0)", loginUsernameForRegistration)
			conflictMsg = fmt.Sprintf("第 %d 行: 用户名 '%s' 已在此平台注册。", rowIndex, loginUsernameForRegistration)
		} else if currentEmailAccountIDPtr != nil { // loginIdentifier was an email
			conflictQuery = conflictQuery.Where("(login_username = '' OR login_username IS NULL) AND email_account_id = ?", *currentEmailAccountIDPtr)
			conflictMsg = fmt.Sprintf("第 %d 行: 邮箱 '%s' (ID: %d) 已在此平台注册。", rowIndex, loginIdentifier, *currentEmailAccountIDPtr)
		} else {
			// Should have been caught by empty loginIdentifier check earlier
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第 %d 行 (平台 %s): 内部错误，无有效登录标识符进行冲突检查", rowIndex, platform.Name))
			log.Printf("Import: Row %d internal error, no valid login identifier for conflict check (Platform: %s)", rowIndex, platform.Name)
			continue
		}

		errLoop = conflictQuery.First(&existingRegistration).Error
		if errLoop == nil { // Record found, means conflict with an existing registration
			errorCount++
			errorMessages = append(errorMessages, conflictMsg)
			log.Printf("Import: Row %d conflict: %s", rowIndex, conflictMsg)
			continue
		} else if !errors.Is(errLoop, gorm.ErrRecordNotFound) { // Actual DB error during conflict check
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第 %d 行: 检查平台注册冲突失败: %v", rowIndex, errLoop))
			log.Printf("Import: Row %d error checking registration conflict: %v", rowIndex, errLoop)
			continue
		}
		// No conflicting registration found.

		// 5. Create PlatformRegistration
		combinedNotes := item.Notes
		if len(item.CustomFields) > 0 {
			customFieldsJSON, jsonErr := json.Marshal(item.CustomFields)
			if jsonErr == nil {
				combinedNotes += "\n\n--- Custom Fields (Imported) ---\n" + string(customFieldsJSON)
			} else {
				combinedNotes += "\n\n--- Custom Fields (Raw, JSON marshal error) ---\n"
				for k, vCustom := range item.CustomFields {
					combinedNotes += fmt.Sprintf("%s: %s\n", k, vCustom)
				}
				log.Printf("Import: Row %d warning, could not marshal custom fields for login '%s', platform '%s': %v", rowIndex, loginIdentifier, platform.Name, jsonErr)
			}
		}

		var loginUsernamePtr *string
		if loginUsernameForRegistration != "" {
			loginUsernamePtr = &loginUsernameForRegistration
		}

		registration := models.PlatformRegistration{
			UserID:                 userID,
			EmailAccountID:         currentEmailAccountIDPtr,
			PlatformID:             platform.ID,
			LoginUsername:          loginUsernamePtr,
			LoginPasswordEncrypted: encryptedPassword,
			Notes:                  combinedNotes,
			PhoneNumber:            "", // Bitwarden CSV doesn't map directly to this.
		}

		if createErr := db.Create(&registration).Error; createErr != nil {
			errorCount++
			errMsg := fmt.Sprintf("第 %d 行 (平台 %s, 登录名 '%s'): 创建平台注册信息失败: %v", rowIndex, platform.Name, loginIdentifier, createErr)
			if strings.Contains(createErr.Error(), "UNIQUE constraint failed") || strings.Contains(createErr.Error(), "UNIQUE constraint violation") {
				errMsg = fmt.Sprintf("第 %d 行 (平台 %s, 登录名 '%s'): 创建失败，唯一约束冲突 (可能已存在): %v", rowIndex, platform.Name, loginIdentifier, createErr)
			}
			errorMessages = append(errorMessages, errMsg)
			log.Printf("Import: Row %d error creating registration (Platform: %s, Login: '%s'): %v", rowIndex, platform.Name, loginIdentifier, createErr)
			continue
		} else {
			savedCount++
			log.Printf("Import: Row %d successfully created registration ID %d for login '%s', platform '%s'", rowIndex, registration.ID, loginIdentifier, platform.Name)
		}
	}
	// --- 数据库保存逻辑结束 ---

	log.Printf("Import finished. Saved: %d, Errors: %d", savedCount, errorCount)

	// 根据保存结果返回响应
	responseMessage := fmt.Sprintf("Bitwarden CSV 文件处理完成。成功保存 %d 条记录。", savedCount)
	if errorCount > 0 {
		responseMessage += fmt.Sprintf(" 遇到 %d 个错误。", errorCount)
	}

	// 决定状态码：如果完全没有保存成功，可能返回错误码？或者总是返回200但包含错误信息？
	// 暂时总是返回 200 OK，让前端根据 savedCount 和 errorMessages 显示详情
	// finalStatusCode := http.StatusOK // 移除未使用的变量

	utils.SendSuccessResponse(c, gin.H{
		"message":       responseMessage,
		"savedCount":    savedCount, // 使用 savedCount 替代 importedCount
		"errorCount":    errorCount,
		"errorMessages": errorMessages, // 返回具体的错误信息列表
		// "items": items, // 不再返回原始解析项，减少响应大小
	})
}
