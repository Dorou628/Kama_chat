# KamaChat 实时聊天系统

<div align="center">
  <h3>🚀 基于 Go + Vue3 的现代化即时通讯系统</h3>
  <p>一个功能完整的仿微信聊天应用，支持单聊、群聊、音视频通话等功能</p>
</div>

## 📖 项目介绍

KamaChat 是一个前后端分离的即时通讯项目，采用现代化的技术栈构建，具备完整的聊天功能和用户管理系统。项目旨在打造类似微信的聊天体验，支持多种消息类型、实时通讯、音视频通话等核心功能。

### ✨ 核心特性

- 🔐 **安全认证**: SMS短信验证登录，SSL加密传输
- 💬 **即时通讯**: WebSocket实时消息推送，支持单聊和群聊
- 👥 **联系人管理**: 添加好友、群组管理、黑名单功能
- 📁 **多媒体消息**: 支持文本、图片、文件、音视频消息
- 📞 **音视频通话**: 基于WebRTC的1对1音视频通话
- 📱 **离线消息**: 消息持久化存储，离线消息推送
- 🎛️ **后台管理**: 用户管理、系统监控等管理功能
- 🚀 **高性能**: Redis缓存、Kafka消息队列、数据库优化

## 🏗️ 技术栈

### 后端技术
- **语言**: Go 1.20+
- **框架**: Gin (Web框架)
- **数据库**: MySQL + GORM (ORM)
- **缓存**: Redis
- **消息队列**: Kafka (可选)
- **WebSocket**: Gorilla WebSocket
- **日志**: Zap
- **短信服务**: 阿里云SMS

### 前端技术
- **框架**: Vue 3.2+
- **UI组件**: Element Plus
- **状态管理**: Vuex 4
- **路由**: Vue Router 4
- **HTTP客户端**: Axios
- **实时通讯**: WebSocket
- **音视频**: WebRTC

## 📁 项目结构

```
KamaChat/
├── 📁 api/v1/                    # API控制器
│   ├── chatroom_controller.go    # 聊天室控制器
│   ├── message_controller.go     # 消息控制器
│   ├── user_info_controller.go   # 用户信息控制器
│   └── ws_controller.go          # WebSocket控制器
├── 📁 cmd/kama_chat_server/      # 应用入口
│   └── main.go                   # 主程序
├── 📁 configs/                   # 配置文件
│   └── config.toml               # 主配置文件
├── 📁 internal/                  # 内部模块
│   ├── 📁 config/                # 配置管理
│   ├── 📁 dao/                   # 数据访问层
│   ├── 📁 model/                 # 数据模型
│   └── 📁 service/               # 业务服务层
│       ├── 📁 chat/              # 聊天服务
│       ├── 📁 redis/             # Redis服务
│       ├── 📁 kafka/             # Kafka服务
│       └── 📁 sms/               # 短信服务
├── 📁 pkg/                       # 公共包
│   ├── 📁 constants/             # 常量定义
│   ├── 📁 util/                  # 工具函数
│   └── 📁 zlog/                  # 日志工具
├── 📁 web/chat-server/           # 前端项目
│   ├── 📁 src/                   # 源代码
│   │   ├── 📁 components/        # Vue组件
│   │   ├── 📁 views/             # 页面视图
│   │   ├── 📁 store/             # Vuex状态管理
│   │   └── 📁 router/            # 路由配置
│   ├── package.json              # 前端依赖
│   └── vue.config.js             # Vue配置
├── 📁 static/                    # 静态资源
│   ├── 📁 avatars/               # 用户头像
│   └── 📁 files/                 # 文件存储
├── start_services.bat            # Windows启动脚本
├── start_services.ps1            # PowerShell启动脚本
└── stop_services.ps1             # 停止服务脚本
```

## 🛠️ 环境要求

### 系统要求
- **操作系统**: Windows 10/11, macOS, Linux
- **Go版本**: 1.20 或更高
- **Node.js版本**: 16.0 或更高
- **数据库**: MySQL 8.0 或更高
- **缓存**: Redis 6.0 或更高

### 可选组件
- **Kafka**: 2.8+ (用于消息队列，可选)
- **SSL证书**: 用于HTTPS部署

## 📦 安装依赖

### 1. 后端依赖安装

```bash
# 克隆项目
git clone https://github.com/Dorou628/Kama_chat.git
cd KamaChat

# 安装Go依赖
go mod download
go mod tidy
```

### 2. 前端依赖安装

```bash
# 进入前端目录
cd web/chat-server

# 安装Node.js依赖
npm install
# 或使用yarn
yarn install
```

### 3. 数据库安装

#### MySQL安装
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install mysql-server

# CentOS/RHEL
sudo yum install mysql-server

# Windows
# 下载MySQL安装包: https://dev.mysql.com/downloads/mysql/
```

#### Redis安装
```bash
# Ubuntu/Debian
sudo apt install redis-server

# CentOS/RHEL
sudo yum install redis

# Windows
# 下载Redis: https://github.com/microsoftarchive/redis/releases
```

## ⚙️ 配置说明

### 1. 数据库配置

创建MySQL数据库：
```sql
CREATE DATABASE kama_chat CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 配置文件修改

编辑 `configs/config.toml` 文件：

```toml
[mainConfig]
appName = "KamaChat"
host = "0.0.0.0"
port = 8000

[mysqlConfig]
host = "127.0.0.1"
port = 3306
user = "root"
password = "your_mysql_password"        # 修改为你的MySQL密码
databaseName = "kama_chat"

[redisConfig]
host = "127.0.0.1"
port = 6379
password = ""                           # 如果Redis设置了密码，请填写
db = 0

[authCodeConfig]
accessKeyID = "PLACEHOLDER_ACCESS_KEY_ID"      # 替换为你的阿里云AccessKey ID
accessKeySecret = "PLACEHOLDER_ACCESS_KEY_SECRET"  # 替换为你的阿里云AccessKey Secret
signName = "阿里云短信测试"
templateCode = "SMS_154950909"

[smsAuthConfig]
accessKeyID = "PLACEHOLDER_ACCESS_KEY_ID"      # 替换为你的阿里云AccessKey ID
accessKeySecret = "PLACEHOLDER_ACCESS_KEY_SECRET"  # 替换为你的阿里云AccessKey Secret
useNewService = true

[logConfig]
logPath = "./logs"                      # 日志文件路径

[kafkaConfig]
messageMode = "channel"                 # 消息模式: channel/kafka/hybrid
hostPort = "127.0.0.1:9092"
loginTopic = "login"
chatTopic = "chat_message"
logoutTopic = "logout"
partition = 0
timeout = 1

[staticSrcConfig]
staticAvatarPath = "./static/avatars"
staticFilePath = "./static/files"
```

### 3. 阿里云短信服务配置

1. 登录阿里云控制台
2. 开通短信服务
3. 获取AccessKey ID和AccessKey Secret
4. 申请短信签名和模板
5. 将相关信息填入配置文件

## 🚀 启动方法

### 方式一：一键启动（推荐）

#### Windows用户
```bash
# 双击运行批处理文件
start_services.bat

# 或在命令行中运行
.\start_services.bat
```

#### PowerShell用户
```powershell
# 运行PowerShell脚本
.\start_services.ps1
```

### 方式二：手动启动

#### 1. 启动Redis
```bash
# Windows
C:\Redis\redis-server.exe C:\Redis\redis.windows.conf

# Linux/macOS
redis-server
```

#### 2. 启动后端服务
```bash
# 在项目根目录下
go run cmd/kama_chat_server/main.go
```

#### 3. 启动前端服务
```bash
# 进入前端目录
cd web/chat-server

# 启动开发服务器
npm run serve
# 或
yarn serve
```

### 停止服务
```powershell
# 运行停止脚本
.\stop_services.ps1
```

## 🌐 访问地址

启动成功后，可以通过以下地址访问：

- **前端应用**: http://localhost:8080
- **后端API**: http://localhost:8000
- **WebSocket**: ws://localhost:8000/ws

## 📱 功能说明

### 用户功能
- ✅ 手机号注册/登录
- ✅ 个人信息管理
- ✅ 头像上传
- ✅ 好友添加/删除
- ✅ 群组创建/管理
- ✅ 单聊/群聊
- ✅ 文件传输
- ✅ 音视频通话
- ✅ 消息历史记录
- ✅ 离线消息

### 管理功能
- ✅ 用户管理
- ✅ 群组管理
- ✅ 消息监控
- ✅ 系统日志
- ✅ 数据统计

## 🔧 开发说明

### 消息模式配置
项目支持三种消息处理模式：

1. **channel模式**: 使用Go channel处理消息（默认）
2. **kafka模式**: 使用Kafka消息队列
3. **hybrid模式**: 混合模式，根据负载自动切换

### 音视频通话配置
如需支持公网音视频通话，需要配置TURN服务器：

```javascript
// 前端配置文件中修改ICE_CFG
const ICE_CFG = {
  iceServers: [
    {
      urls: 'turn:your-turn-server:3478',
      username: 'your-username',
      credential: 'your-password'
    }
  ]
};
```

## 🐛 常见问题

### 1. 数据库连接失败
- 检查MySQL服务是否启动
- 确认数据库用户名密码正确
- 检查数据库是否已创建

### 2. Redis连接失败
- 检查Redis服务是否启动
- 确认Redis配置正确

### 3. 短信发送失败
- 检查阿里云AccessKey配置
- 确认短信模板和签名已审核通过
- 检查账户余额

### 4. 前端无法连接后端
- 检查后端服务是否启动
- 确认端口配置正确
- 检查防火墙设置

## 📄 许可证

本项目采用 MIT 许可证，详情请参阅 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进项目！

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- GitHub Issues: [提交问题](https://github.com/Dorou628/Kama_chat/issues)
- 项目地址: https://github.com/Dorou628/Kama_chat

---

<div align="center">
  <p>⭐ 如果这个项目对你有帮助，请给个Star支持一下！</p>
</div>