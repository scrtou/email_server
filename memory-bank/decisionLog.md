# Decision Log

This file records architectural and implementation decisions using a list format.
2025-06-08 11:09:58 - Log of updates made.

*
      
## Decision

*
      
## Rationale 

*

## Implementation Details

*
---
### Decision
[2025-06-08 11:10:11] - 采用模块化设计，将 IMAP 客户端逻辑封装在专用的 Go 服务中，以实现关注点分离和可维护性。

**Rationale:**
将 IMAP 协议的复杂性与核心业务逻辑（API 处理、用户管理）分离开来，可以使代码更清晰、更易于测试和独立更新。如果未来需要支持其他邮件协议（如 POP3），这种模块化方法也更易于扩展。

**Implications/Details:**
- 将在 `src/backend/` 中创建一个新的 `imap_client` 包。
- 该服务将负责处理与外部 IMAP 服务器的所有连接、认证和数据获取操作。
- API 处理器 (`src/backend/handlers/email.go`) 将调用此服务，而不是直接处理 IMAP 连接。

---
### Decision
[2025-06-08 11:10:11] - 使用第三方库 `github.com/emersion/go-imap` 来处理 IMAP 通信。

**Rationale:**
从头开始实现完整的 IMAP 客户端既复杂又耗时，且容易出错。`go-imap` 是一个成熟、功能齐全且维护良好的库，它抽象了 IMAP 协议的底层细节，使我们能够专注于业务逻辑。

**Implications/Details:**
- 需要将 `github.com/emersion/go-imap` 添加到 `go.mod` 文件中。
- `imap_client` 服务将使用这个库来连接、验证和获取邮件。

---
### Decision
[2025-06-08 11:10:11] - 对用户邮箱账户凭据（密码/应用密码）进行加密存储。

**Rationale:**
以明文形式存储敏感的用户凭据是重大的安全风险。必须使用强大的加密措施来保护这些数据，防止未经授权的访问。

**Implications/Details:**
- 将利用项目现有的 `src/backend/utils/encryption.go` 中的加密/解密函数。
- 在将凭据存入数据库之前，`email_account` 处理器将对其进行加密。
- 在需要连接 IMAP 服务器之前，`imap_client` 服务将请求解密凭据。
---
### Decision
[2025-06-08 14:49:34] - Adopt the standard Go package `golang.org/x/oauth2` for handling the client-side OAuth2 flow.

**Rationale:**
Using the official, well-maintained, and widely adopted `golang.org/x/oauth2` package is the industry standard for Go applications. It provides a robust, secure, and comprehensive implementation of the OAuth2 client specification, saving significant development time and reducing the risk of security flaws compared to a manual implementation. It supports various grant types, including the Authorization Code Grant flow required for this integration.

**Implications/Details:**
- The package will be added as a dependency in the `go.mod` file.
- New handlers in `src/backend/handlers/oauth2.go` will use this package to:
  - Generate authorization request URLs.
  - Exchange authorization codes for access and refresh tokens.
  - Manage and refresh tokens automatically.
- A new configuration section will be required to store OAuth2 client IDs and secrets for each supported provider.