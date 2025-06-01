package models

// ImportedLoginItem 代表从外部源导入的登录条目的统一数据模型
type ImportedLoginItem struct {
	SourceName   string            `json:"source_name"`   // 数据来源，例如："Bitwarden"
	ItemName     string            `json:"item_name"`     // 条目名称
	Username     string            `json:"username"`      // 用户名
	Password     string            `json:"password"`      // 密码 (可选，根据用户选择导入)
	URL          string            `json:"url"`           // 相关 URL
	Notes        string            `json:"notes"`         // 备注信息
	Folder       string            `json:"folder"`        // 文件夹名称 (可选)
	TOTP         string            `json:"totp"`          // TOTP 密钥 (可选)
	CustomFields map[string]string `json:"custom_fields"` // 自定义字段 (可选)
}