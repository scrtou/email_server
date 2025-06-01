# Bitwarden CSV 导入功能设计方案

## 1. 统一数据模型定义 (`ImportedLoginItem` Go 结构体)

我们将定义一个名为 `ImportedLoginItem` 的 Go 结构体，用于在内部表示导入的登录信息。

```go
package models // 或其他合适的包名

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
```

**字段用途说明:**

*   `SourceName`: (string) 标识数据来源，对于 Bitwarden CSV，固定为 `"Bitwarden"`。
*   `ItemName`: (string) 对应 CSV 中的 `name` 列，是登录条目的主要名称。
*   `Username`: (string) 对应 CSV 中的 `login_username` 列。
*   `Password`: (string) 对应 CSV 中的 `login_password` 列。此字段的填充将取决于用户是否选择导入密码。
*   `URL`: (string) 对应 CSV 中的 `login_uri` 列。
*   `Notes`: (string) 对应 CSV 中的 `notes` 列。
*   `Folder`: (string) 对应 CSV 中的 `folder` 列。
*   `TOTP`: (string) 对应 CSV 中的 `login_totp` 列。
*   `CustomFields`: (map[string]string) 用于存储未能直接映射到上述标准字段的其他数据。
    *   CSV 中的 `reprompt` 列将存储为 `CustomFields["reprompt"]`。
    *   CSV 中的 `favorite` 列将存储为 `CustomFields["favorite"]`。
    *   CSV 中的 `type` 列将存储为 `CustomFields["type"]`。
    *   CSV 中的 `fields` 列（如果包含有效的 JSON 数据）的内容也将合并到此 map 中。

## 2. Bitwarden CSV 解析与映射策略

**CSV 列:** `folder`, `favorite`, `type`, `name`, `notes`, `fields`, `reprompt`, `login_uri`, `login_username`, `login_password`, `login_totp`

**映射规则:**

| Bitwarden CSV 列 | `ImportedLoginItem` 字段 | 处理说明                                                                                                                                                                                             |
| :--------------- | :----------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `folder`         | `Folder`                 | 直接映射。                                                                                                                                                                                             |
| `favorite`       | `CustomFields["favorite"]` | 存储为自定义字段。                                                                                                                                                                                     |
| `type`           | `CustomFields["type"]`   | 存储为自定义字段。                                                                                                                                                                                     |
| `name`           | `ItemName`               | 直接映射。                                                                                                                                                                                             |
| `notes`          | `Notes`                  | 直接映射。                                                                                                                                                                                             |
| `fields`         | `CustomFields`           | **特殊处理**: 若此列包含 JSON 字符串，则解析 JSON 并将其中的键值对添加到 `CustomFields`。若为空或无效 JSON，则可记录警告或忽略（根据示例，此列可能为空）。                                                               |
| `reprompt`       | `CustomFields["reprompt"]` | 存储为自定义字段，键为 "reprompt"。                                                                                                                                                                    |
| `login_uri`      | `URL`                    | 直接映射。                                                                                                                                                                                             |
| `login_username` | `Username`               | 直接映射。                                                                                                                                                                                             |
| `login_password` | `Password`               | **条件映射**: 根据用户“是否导入密码”的选择进行填充。若不导入，则此字段为空字符串。                                                                                                                            |
| `login_totp`     | `TOTP`                   | 直接映射。                                                                                                                                                                                             |

**处理 `fields` 列的逻辑:**
*   读取 `fields` 列的字符串。
*   如果非空，尝试将其作为 JSON 解析。
*   如果解析成功，将 JSON 对象中的每个键值对添加到 `ImportedLoginItem.CustomFields`。
*   如果解析失败或为空，可以记录警告。鉴于您提供的示例中此列为空，我们将主要确保在它有内容时能正确处理。

**处理密码导入选项:**
*   应用程序将提供一个选项，让用户决定是否导入密码。
*   如果用户选择导入，`login_password` 列的数据将填充到 `ImportedLoginItem.Password`。
*   否则，`ImportedLoginItem.Password` 将为空字符串。

## 3. Mermaid 图 (可视化映射关系)

```mermaid
graph LR
    subgraph Bitwarden CSV Row
        csv_folder["folder"]
        csv_favorite["favorite"]
        csv_type["type"]
        csv_name["name"]
        csv_notes["notes"]
        csv_fields["fields (JSON)"]
        csv_reprompt["reprompt"]
        csv_login_uri["login_uri"]
        csv_login_username["login_username"]
        csv_login_password["login_password"]
        csv_login_totp["login_totp"]
    end

    subgraph ImportedLoginItem Struct
        struct_SourceName["SourceName (='Bitwarden')"]
        struct_ItemName["ItemName"]
        struct_Username["Username"]
        struct_Password["Password (Conditional)"]
        struct_URL["URL"]
        struct_Notes["Notes"]
        struct_Folder["Folder"]
        struct_TOTP["TOTP"]
        struct_CustomFields["CustomFields (map)"]
    end

    csv_folder --> struct_Folder
    csv_name --> struct_ItemName
    csv_login_username --> struct_Username
    csv_login_password --> struct_Password
    csv_login_uri --> struct_URL
    csv_notes --> struct_Notes
    csv_login_totp --> struct_TOTP

    csv_favorite --> struct_CustomFields
    csv_type --> struct_CustomFields
    csv_reprompt --> struct_CustomFields
    csv_fields --> struct_CustomFields

    style "Bitwarden CSV Row" fill:#f9f,stroke:#333,stroke-width:2px
    style "ImportedLoginItem Struct" fill:#ccf,stroke:#333,stroke-width:2px