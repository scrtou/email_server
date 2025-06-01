package importer

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"email_server/models" // 使用相对于模块根目录的路径
)

// ParseBitwardenCSV 解析 Bitwarden 导出的 CSV 数据。
// reader: 包含 CSV 数据。
// importPasswords: 指示是否应导入密码字段。
// 返回解析后的登录条目列表和任何发生的错误。
func ParseBitwardenCSV(reader io.Reader, importPasswords bool) ([]models.ImportedLoginItem, error) {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true // 允许字段内未转义的引号
	// Bitwarden CSV 可能包含不同数量的字段，先不强制字段数量
	// csvReader.FieldsPerRecord = -1 // 允许多变字段数

	// 读取所有记录
	records, err := csvReader.ReadAll()
	if err != nil {
		// 尝试处理可能的引号问题
		if strings.Contains(err.Error(), "wrong number of fields") {
			fmt.Println("警告: CSV 行字段数不匹配，尝试逐行读取...")
			// 回到开头重新读取
			if seeker, ok := reader.(io.Seeker); ok {
				_, seekErr := seeker.Seek(0, io.SeekStart)
				if seekErr != nil {
					return nil, fmt.Errorf("无法重置 reader 以逐行读取: %w", seekErr)
				}
				csvReader = csv.NewReader(reader) // 创建新的 Reader
				csvReader.LazyQuotes = true       // 确保重试时也启用 LazyQuotes
				csvReader.FieldsPerRecord = -1    // 允许变化的字段数
				records, err = csvReader.ReadAll() // 再次尝试读取
				if err != nil {
					return nil, fmt.Errorf("逐行读取 CSV 数据时仍出错: %w", err)
				}
			} else {
				return nil, fmt.Errorf("读取 CSV 数据时出错 (且无法重置 reader): %w", err)
			}
		} else {
			return nil, fmt.Errorf("读取 CSV 数据时出错: %w", err)
		}
	}


	if len(records) < 1 {
		return nil, fmt.Errorf("CSV 文件为空或没有表头")
	}

	// 解析表头以确定列索引
	header := records[0]
	colIndex := make(map[string]int)
	for i, h := range header {
		// 转换为小写并去除空格以便健壮匹配
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// 验证是否包含必要的列（可以根据需要调整哪些是绝对必要的）
	requiredHeaders := []string{"name", "login_username"} // 至少需要名称和用户名
	for _, reqH := range requiredHeaders {
		if _, ok := colIndex[reqH]; !ok {
			// 暂时改为警告，允许继续处理，但可能导致条目不完整
			fmt.Printf("警告: CSV 文件缺少推荐的列: %s\n", reqH)
		}
	}

	var importedItems []models.ImportedLoginItem
	// 从第二行开始处理数据行 (索引 1)
	for i, row := range records[1:] {
		// 检查行是否完全为空或只包含空格
		isEmptyRow := true
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				isEmptyRow = false
				break
			}
		}
		if isEmptyRow {
			fmt.Printf("警告: 跳过第 %d 行，因为它是空行。\n", i+2)
			continue
		}


		item := models.ImportedLoginItem{
			SourceName:   "Bitwarden",
			CustomFields: make(map[string]string),
		}

		// 辅助函数，安全地从行中获取值
		getValue := func(colName string) string {
			lowercaseColName := strings.ToLower(colName)
			if idx, ok := colIndex[lowercaseColName]; ok {
				if idx < len(row) {
					return row[idx]
				}
				// 列存在但当前行数据不足
				fmt.Printf("警告: 第 %d 行缺少列 '%s' 的数据 (索引 %d 超出范围 %d)\n", i+2, colName, idx, len(row))
				return ""
			}
			// 列名在表头中不存在
			// fmt.Printf("调试: 列名 '%s' 在表头中未找到\n", colName) // 可选的调试信息
			return ""
		}

		// 应用映射规则
		item.Folder = getValue("folder")
		item.ItemName = getValue("name")
		item.Notes = getValue("notes")
		item.URL = getValue("login_uri")
		item.Username = getValue("login_username")
		item.TOTP = getValue("login_totp")

		// 条件映射密码
		if importPasswords {
			item.Password = getValue("login_password")
		} else {
			item.Password = "" // 明确设置为空
		}

		// 映射到 CustomFields
		if fav := getValue("favorite"); fav != "" {
			item.CustomFields["favorite"] = fav
		}
		if typ := getValue("type"); typ != "" {
			item.CustomFields["type"] = typ
		}
		if reprompt := getValue("reprompt"); reprompt != "" {
			item.CustomFields["reprompt"] = reprompt
		}

		// 处理 fields 列 (JSON)
		fieldsData := getValue("fields")
		if fieldsData != "" {
			var customFieldsFromJson map[string]interface{} // 使用 interface{} 以处理不同类型的值
			// 尝试去除可能的外部引号（如果 fields 数据被额外引用）
			unquotedData := strings.Trim(fieldsData, `"`)
			err := json.Unmarshal([]byte(unquotedData), &customFieldsFromJson)
			if err != nil {
				// 解析失败，记录警告并将原始数据存储在 _raw_fields_data 中
				fmt.Printf("警告: 解析第 %d 行的 'fields' 列 JSON 时出错: %v。将原始数据存储在 _raw_fields_data 中。原始数据: %s\n", i+2, err, fieldsData)
				item.CustomFields["_raw_fields_data"] = fieldsData // 总是存储原始数据以供调试
			} else {
				// 解析成功
				for key, value := range customFieldsFromJson {
					// 将解析出的值转换为字符串存储
					// 避免覆盖已经从其他列映射过来的字段（如 favorite, type, reprompt）
					lowercaseKey := strings.ToLower(key)
					if _, exists := item.CustomFields[lowercaseKey]; !exists {
						item.CustomFields[key] = fmt.Sprintf("%v", value)
					} else {
						// 如果键冲突，可以添加前缀或后缀
						item.CustomFields[fmt.Sprintf("field_%s", key)] = fmt.Sprintf("%v", value)
						fmt.Printf("警告: 第 %d 行 'fields' 列中的键 '%s' 与其他列冲突，已重命名存储。\n", i+2, key)
					}
				}
			}
		}

		// 只有当条目名称或用户名至少有一个不为空时才添加，避免大部分空的条目
		if item.ItemName != "" || item.Username != "" {
			importedItems = append(importedItems, item)
		} else {
			fmt.Printf("警告: 跳过第 %d 行，因为名称和用户名都为空。\n", i+2)
		}
	}

	return importedItems, nil
}