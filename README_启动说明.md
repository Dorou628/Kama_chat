# KamaChat 一键启动脚本使用说明

## 📋 概述

本项目提供了多种一键启动脚本，方便快速启动KamaChat的所有服务组件。

## 🚀 启动脚本

### 1. Windows批处理脚本 (推荐新手使用)

**文件**: `start_services.bat`

**使用方法**:
- 双击 `start_services.bat` 文件
- 或在命令行中运行: `start_services.bat`

**特点**:
- 简单易用，双击即可运行
- 自动检测Redis状态
- 分别在新窗口启动各个服务
- 提供详细的启动信息

### 2. PowerShell脚本 (推荐高级用户)

**文件**: `start_services.ps1`

**使用方法**:
- 右键点击 `start_services.ps1` → "使用PowerShell运行"
- 或在PowerShell中运行: `.\start_services.ps1`

**特点**:
- 更强大的错误处理
- 端口检测和等待机制
- 彩色输出，更好的用户体验
- 自动检测服务是否已在运行

## 🛑 停止服务

**文件**: `stop_services.ps1`

**使用方法**:
- 右键点击 `stop_services.ps1` → "使用PowerShell运行"
- 或在PowerShell中运行: `.\stop_services.ps1`

**功能**:
- 自动停止所有KamaChat相关服务
- 根据端口和进程名清理残留进程
- 确保完全停止所有服务

## 📦 启动的服务

脚本会按顺序启动以下服务：

1. **Redis服务** (端口: 6379)
   - 路径: `C:\Redis\redis-server.exe`
   - 配置: `C:\Redis\redis.windows.conf`

2. **Go后端服务** (端口: 8000)
   - 命令: `go run cmd/kama_chat_server/main.go`
   - API接口: `http://localhost:8000`
   - WebSocket: `ws://localhost:8000/wss`

3. **Vue前端服务** (端口: 8080)
   - 命令: `npm run serve`
   - 访问地址: `http://localhost:8080`

## ⚙️ 前置要求

确保以下软件已正确安装：

- **Redis**: 安装在 `C:\Redis\` 目录
- **Go**: 版本 1.19 或更高
- **Node.js**: 版本 14 或更高
- **npm**: 随Node.js一起安装

## 🔧 故障排除

### Redis启动失败
- 检查Redis是否安装在 `C:\Redis\` 目录
- 确保 `redis-server.exe` 和 `redis.windows.conf` 文件存在
- 检查端口6379是否被其他程序占用

### Go后端启动失败
- 确保在项目根目录运行脚本
- 检查Go环境是否正确配置
- 运行 `go mod tidy` 确保依赖完整

### Vue前端启动失败
- 确保已运行 `npm install` 安装依赖
- 检查Node.js和npm版本
- 检查端口8080是否被占用

## 📝 注意事项

1. **首次运行**: 确保已安装所有依赖
2. **端口冲突**: 如果端口被占用，请先停止占用端口的程序
3. **权限问题**: 某些情况下可能需要以管理员身份运行
4. **防火墙**: 确保防火墙允许相关端口的访问

## 🎯 快速开始

1. 确保所有前置要求已满足
2. 双击运行 `start_services.bat`
3. 等待所有服务启动完成
4. 打开浏览器访问 `http://localhost:8080`
5. 开始使用KamaChat！

## 📞 技术支持

如果遇到问题，请检查：
1. 各个服务的终端窗口是否有错误信息
2. 端口是否被正确占用
3. 依赖是否完整安装

---

**祝您使用愉快！** 🎉