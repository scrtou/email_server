# 部署文档 - Email Server 应用

## 简介

本文档旨在指导您如何部署 Email Server 应用。Email Server 应用是一个用于管理邮件账户、平台注册信息和相关服务的Web应用程序。本文档将涵盖从获取代码到在本地环境中使用 Docker Compose 启动应用的完整步骤，并提供生产环境部署的一些建议。

## 先决条件

在开始部署之前，请确保您的系统已安装以下软件和工具：

*   **Git**: 用于从版本控制系统获取源代码。
*   **Docker**: 用于容器化应用。
*   **Docker Compose**: 用于定义和运行多容器 Docker 应用程序。

## 获取代码

首先，您需要从版本控制系统（例如 Git）克隆最新的项目代码到您的本地计算机。

```bash
git clone <repository_url>
cd email_server
```
请将 `<repository_url>` 替换为实际的代码仓库地址。

## 环境配置

应用的环境配置主要通过 Docker Compose 文件进行管理。

### 后端服务环境变量

后端服务所需的环境变量在 [`docker-compose.yml`](docker-compose.yml:1) 文件中定义。关键变量包括：

*   `SERVER_PORT`: 后端服务在容器内监听的端口，默认为 `5555`。
*   `SQLITE_FILE`: SQLite 数据库文件在容器内的路径，默认为 `/data/database.db`。
*   `JWT_SECRET`: 用于 JWT 签名的密钥。**重要提示：请务必在生产环境中修改此密钥为一个强随机字符串。**

您可以根据需要在 [`docker-compose.yml`](docker-compose.yml:1) 文件的 `backend.environment` 部分添加或修改其他环境变量。

### 前端服务环境变量

前端应用可能需要配置 API 的基础 URL。这可以通过以下方式之一进行配置：

1.  **构建时参数 (Build Args)**:
    在 [`docker-compose.yml`](docker-compose.yml:1) 文件的 `frontend.build.args` 部分，可以设置 `VUE_APP_API_BASE_URL`。
    例如：
    ```yaml
    services:
      frontend:
        build:
          context: ./src/frontend
          dockerfile: Dockerfile
          args:
            VUE_APP_API_BASE_URL: "http://localhost:5555/api/v1" # 或 "http://backend:5555/api/v1" 当容器间通信时
    ```
    当使用 `http://localhost:5555/api/v1` 时，前端应用会通过宿主机的 `5555` 端口访问后端。如果前端容器直接通过 Docker 网络访问后端容器，可以使用服务名，例如 `http://backend:5555/api/v1` (假设后端服务名为 `backend`，并在 `5555` 端口运行)。

2.  **运行时配置 (如果前端应用支持)**:
    如果您的前端应用支持在运行时通过挂载配置文件或注入环境变量来设置 API 地址，这也是一种选择。但当前 [`docker-compose.yml`](docker-compose.yml:1) 的配置侧重于构建时注入。

### 数据库持久化

SQLite 数据库文件将持久化到宿主机的 `./data/backend` 目录。这是通过 [`docker-compose.yml`](docker-compose.yml:1) 中的 `volumes` 配置实现的：

```yaml
services:
  backend:
    volumes:
      - ./data/backend:/data # 宿主机路径:容器内路径
```
这意味着即使容器停止或删除，数据库文件也会保留在宿主机的 `./data/backend` 目录下。

## 构建与启动应用 (使用 Docker Compose)

使用 Docker Compose 可以方便地构建和管理应用容器。

1.  **构建并后台启动应用**：
    在项目的根目录（包含 [`docker-compose.yml`](docker-compose.yml:1) 文件的目录）下运行以下命令：

    ```bash
    docker-compose up --build -d
    ```
    *   `--build`: 指示 Docker Compose 在启动容器前重新构建镜像。
    *   `-d`: 以分离模式（detached mode）在后台运行容器。

2.  **查看容器日志**：
    要查看正在运行的容器的日志（例如，用于调试），可以使用以下命令：

    ```bash
    docker-compose logs -f
    ```
    *   `-f`: 持续跟踪日志输出。您可以指定服务名称来查看特定服务的日志，例如 `docker-compose logs -f backend`。

3.  **停止应用**：
    要停止并移除由 `docker-compose up` 创建的容器、网络和卷（除非卷被声明为外部卷），请运行：

    ```bash
    docker-compose down
    ```
    如果您只想停止容器而不移除它们，可以使用 `docker-compose stop`。

## 访问应用

应用成功启动后，您可以通过以下 URL 访问：

*   **前端应用**: `http://localhost:8080`
    *   这是通过 [`docker-compose.yml`](docker-compose.yml:1) 中前端服务的端口映射 `8080:80` 实现的，其中 `80` 是前端容器内 Nginx 服务的默认端口。
*   **后端 API**: `http://localhost:5555`
    *   这是通过 [`docker-compose.yml`](docker-compose.yml:1) 中后端服务的端口映射 `5555:5555` 实现的。API 的具体端点将基于此基地址，例如 `http://localhost:5555/api/v1/...`。

## 生产环境部署考虑

在将此应用部署到生产环境时，应考虑以下几点以增强安全性、可靠性和性能：

*   **数据库**:
    *   SQLite 适用于开发和小型应用，但在生产环境中，建议使用更健壮的关系型数据库，如 PostgreSQL 或 MySQL。您需要相应地修改后端配置和 [`docker-compose.yml`](docker-compose.yml:1) 文件。
*   **HTTPS 配置**:
    *   为所有外部流量启用 HTTPS 至关重要。这通常通过在前端部署一个反向代理（如 Nginx 或 Traefik）并配置 SSL/TLS 证书来实现。
*   **安全设置**:
    *   **修改 `JWT_SECRET`**: 确保 [`docker-compose.yml`](docker-compose.yml:1) 或环境变量中的 `JWT_SECRET` 被替换为一个非常强大且唯一的密钥。
    *   **其他安全加固**: 根据应用特性，实施其他安全措施，如输入验证、速率限制、防火墙规则等。
*   **备份策略**:
    *   为您的数据库和任何持久化数据制定并实施定期的备份策略。
*   **反向代理**:
    *   使用专业的反向代理（如 Nginx、Apache、Traefik 或 Caddy）来处理外部流量、负载均衡（如果需要）、SSL 终止和提供静态内容。
    *   前端的 [`Dockerfile`](src/frontend/Dockerfile:1) 和 [`nginx.conf`](src/frontend/nginx.conf:1) 已经包含了一个基本的 Nginx 配置，但在生产环境中可能需要更高级的配置。
*   **资源限制与监控**:
    *   为容器设置合理的资源限制（CPU、内存）。
    *   设置监控和告警系统，以跟踪应用性能和健康状况。
*   **环境变量管理**:
    *   在生产环境中，避免将敏感信息（如密钥）直接写入 [`docker-compose.yml`](docker-compose.yml:1)。使用 `.env` 文件（并将其添加到 `.gitignore`）或更安全的密钥管理系统。

## 故障排查

以下是一些部署过程中可能遇到的常见问题及其解决方法：

*   **端口冲突**:
    *   **问题**: 错误信息提示端口已被占用（例如 `Error starting userland proxy: listen tcp4 0.0.0.0:8080: bind: address already in use`）。
    *   **解决**: 检查是否有其他应用正在使用 [`docker-compose.yml`](docker-compose.yml:1) 中定义的宿主机端口（例如 `8080` 或 `5555`）。您可以停止占用端口的应用，或修改 [`docker-compose.yml`](docker-compose.yml:1) 中的宿主机端口映射。
*   **环境变量配置错误**:
    *   **问题**: 应用启动失败或行为异常，日志中可能包含与配置相关的错误。
    *   **解决**: 仔细检查 [`docker-compose.yml`](docker-compose.yml:1) 中的环境变量（特别是 `JWT_SECRET`、数据库连接信息、`VUE_APP_API_BASE_URL`）是否正确设置。确保前端构建时能正确获取到 API 地址。
*   **Docker 镜像构建失败**:
    *   **问题**: `docker-compose up --build` 命令执行失败，日志中指示 [`Dockerfile`](src/backend/Dockerfile:1) 或 [`Dockerfile`](src/frontend/Dockerfile:1) 中的某个步骤出错。
    *   **解决**: 查看详细的构建日志，定位到出错的指令。检查 [`Dockerfile`](src/backend/Dockerfile:1) 或 [`Dockerfile`](src/frontend/Dockerfile:1) 的语法、文件路径、依赖项安装等。
*   **容器无法连接到其他容器 (例如前端到后端)**:
    *   **问题**: 前端应用无法访问后端 API，浏览器控制台可能显示网络错误。
    *   **解决**:
        *   确保所有服务都在同一个 Docker 网络中 (在 [`docker-compose.yml`](docker-compose.yml:1) 中通过 `networks` 定义和分配)。
        *   如果前端通过服务名访问后端 (例如 `http://backend:5555`)，请确保服务名与 [`docker-compose.yml`](docker-compose.yml:1) 中定义的服务名称一致。
        *   检查后端服务的日志，确认其是否正常启动并监听正确的端口。
*   **数据卷挂载问题**:
    *   **问题**: 数据库文件未按预期持久化，或应用无法访问挂载的卷。
    *   **解决**: 检查 [`docker-compose.yml`](docker-compose.yml:1) 中 `volumes` 定义的宿主机路径和容器内路径是否正确。确保 Docker 有权限读写宿主机上的相关目录。

如果遇到其他问题，请仔细查看 `docker-compose logs` 的输出，通常可以从中找到问题的线索。