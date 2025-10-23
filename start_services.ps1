# KamaChat 服务启动脚本 (PowerShell版本)
# 使用方法: 右键点击 -> 使用PowerShell运行

# 设置控制台编码为UTF-8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "========================================" -ForegroundColor Green
Write-Host "       KamaChat 服务启动脚本" -ForegroundColor Green  
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# 获取脚本所在目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# 函数：检查端口是否被占用
function Test-Port {
    param([int]$Port)
    try {
        $connection = New-Object System.Net.Sockets.TcpClient
        $connection.Connect("localhost", $Port)
        $connection.Close()
        return $true
    }
    catch {
        return $false
    }
}

# 函数：等待端口启动
function Wait-ForPort {
    param([int]$Port, [string]$ServiceName, [int]$TimeoutSeconds = 30)
    
    Write-Host "等待 $ServiceName 启动..." -ForegroundColor Yellow
    $timeout = (Get-Date).AddSeconds($TimeoutSeconds)
    
    while ((Get-Date) -lt $timeout) {
        if (Test-Port -Port $Port) {
            Write-Host "✓ $ServiceName 已成功启动 (端口: $Port)" -ForegroundColor Green
            return $true
        }
        Start-Sleep -Seconds 1
    }
    
    Write-Host "✗ $ServiceName 启动超时" -ForegroundColor Red
    return $false
}

# 1. 检查并启动Redis
Write-Host "[1/4] 检查Redis服务状态..." -ForegroundColor Cyan

if (Test-Port -Port 6379) {
    Write-Host "✓ Redis服务已在运行" -ForegroundColor Green
} else {
    Write-Host "[2/4] 启动Redis服务..." -ForegroundColor Cyan
    
    $RedisPath = "C:\Redis\redis-server.exe"
    $RedisConfig = "C:\Redis\redis.windows.conf"
    
    if (Test-Path $RedisPath) {
        Write-Host "正在启动Redis..." -ForegroundColor Yellow
        
        # 启动Redis进程
        $RedisProcess = Start-Process -FilePath $RedisPath -ArgumentList $RedisConfig -WindowStyle Minimized -PassThru
        
        # 等待Redis启动
        if (Wait-ForPort -Port 6379 -ServiceName "Redis" -TimeoutSeconds 10) {
            Write-Host "✓ Redis服务启动完成" -ForegroundColor Green
        } else {
            Write-Host "✗ Redis启动失败，但继续启动其他服务..." -ForegroundColor Yellow
        }
    } else {
        Write-Host "✗ 错误: 未找到Redis安装路径 $RedisPath" -ForegroundColor Red
        Write-Host "  请确保Redis已正确安装在 C:\Redis\ 目录" -ForegroundColor Yellow
    }
}

Write-Host ""

# 2. 启动Go后端服务
Write-Host "[3/4] 启动Go后端服务..." -ForegroundColor Cyan

if (Test-Port -Port 8000) {
    Write-Host "✓ 后端服务已在运行 (端口: 8000)" -ForegroundColor Green
} else {
    Write-Host "正在启动后端服务..." -ForegroundColor Yellow
    
    # 启动Go后端
    $BackendProcess = Start-Process -FilePath "cmd" -ArgumentList "/k", "cd /d `"$ScriptDir`" && go run cmd/kama_chat_server/main.go" -WindowStyle Normal -PassThru
    
    # 等待后端启动
    if (Wait-ForPort -Port 8000 -ServiceName "Go后端" -TimeoutSeconds 15) {
        Write-Host "✓ 后端服务启动完成" -ForegroundColor Green
    } else {
        Write-Host "✗ 后端服务启动可能失败，请检查终端窗口" -ForegroundColor Yellow
    }
}

Write-Host ""

# 3. 启动Vue前端服务
Write-Host "[4/4] 启动Vue前端服务..." -ForegroundColor Cyan

if (Test-Port -Port 8080) {
    Write-Host "✓ 前端服务已在运行 (端口: 8080)" -ForegroundColor Green
} else {
    Write-Host "正在启动前端服务..." -ForegroundColor Yellow
    
    $FrontendPath = Join-Path $ScriptDir "web\chat-server"
    
    if (Test-Path $FrontendPath) {
        # 启动Vue前端
        $FrontendProcess = Start-Process -FilePath "cmd" -ArgumentList "/k", "cd /d `"$FrontendPath`" && npm run serve" -WindowStyle Normal -PassThru
        
        # 等待前端启动
        if (Wait-ForPort -Port 8080 -ServiceName "Vue前端" -TimeoutSeconds 30) {
            Write-Host "✓ 前端服务启动完成" -ForegroundColor Green
        } else {
            Write-Host "✗ 前端服务启动可能失败，请检查终端窗口" -ForegroundColor Yellow
        }
    } else {
        Write-Host "✗ 错误: 未找到前端目录 $FrontendPath" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "           启动完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

Write-Host "服务访问地址:" -ForegroundColor White
Write-Host "  前端应用: " -NoNewline -ForegroundColor White
Write-Host "http://localhost:8080" -ForegroundColor Cyan
Write-Host "  后端API:  " -NoNewline -ForegroundColor White  
Write-Host "http://localhost:8000" -ForegroundColor Cyan
Write-Host "  WebSocket: " -NoNewline -ForegroundColor White
Write-Host "ws://localhost:8000/wss" -ForegroundColor Cyan

Write-Host ""
Write-Host "注意:" -ForegroundColor Yellow
Write-Host "- 所有服务已在独立窗口中启动" -ForegroundColor White
Write-Host "- 如需停止服务，请关闭对应的命令行窗口" -ForegroundColor White
Write-Host "- 或使用 stop_services.ps1 脚本停止所有服务" -ForegroundColor White

Write-Host ""
Write-Host "按任意键退出..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")