# KamaChat 服务停止脚本 (PowerShell版本)
# 使用方法: 右键点击 -> 使用PowerShell运行

# 设置控制台编码为UTF-8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "========================================" -ForegroundColor Red
Write-Host "       KamaChat 服务停止脚本" -ForegroundColor Red
Write-Host "========================================" -ForegroundColor Red
Write-Host ""

# 函数：根据端口杀死进程
function Stop-ProcessByPort {
    param([int]$Port, [string]$ServiceName)
    
    try {
        # 查找占用端口的进程
        $netstatOutput = netstat -ano | Select-String ":$Port "
        
        if ($netstatOutput) {
            foreach ($line in $netstatOutput) {
                # 提取PID
                $pid = ($line -split '\s+')[-1]
                
                if ($pid -and $pid -match '^\d+$') {
                    $process = Get-Process -Id $pid -ErrorAction SilentlyContinue
                    
                    if ($process) {
                        Write-Host "正在停止 $ServiceName (PID: $pid, 进程: $($process.ProcessName))..." -ForegroundColor Yellow
                        Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
                        Write-Host "✓ $ServiceName 已停止" -ForegroundColor Green
                    }
                }
            }
        } else {
            Write-Host "✓ $ServiceName 未在运行 (端口 $Port)" -ForegroundColor Gray
        }
    }
    catch {
        Write-Host "✗ 停止 $ServiceName 时出错: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# 函数：根据进程名杀死进程
function Stop-ProcessByName {
    param([string]$ProcessName, [string]$ServiceName)
    
    try {
        $processes = Get-Process -Name $ProcessName -ErrorAction SilentlyContinue
        
        if ($processes) {
            foreach ($process in $processes) {
                Write-Host "正在停止 $ServiceName (PID: $($process.Id))..." -ForegroundColor Yellow
                Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
            }
            Write-Host "✓ $ServiceName 已停止" -ForegroundColor Green
        } else {
            Write-Host "✓ $ServiceName 未在运行" -ForegroundColor Gray
        }
    }
    catch {
        Write-Host "✗ 停止 $ServiceName 时出错: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "[1/3] 停止Vue前端服务 (端口: 8080)..." -ForegroundColor Cyan
Stop-ProcessByPort -Port 8080 -ServiceName "Vue前端服务"

Write-Host ""
Write-Host "[2/3] 停止Go后端服务 (端口: 8000)..." -ForegroundColor Cyan
Stop-ProcessByPort -Port 8000 -ServiceName "Go后端服务"

Write-Host ""
Write-Host "[3/3] 停止Redis服务 (端口: 6379)..." -ForegroundColor Cyan
Stop-ProcessByPort -Port 6379 -ServiceName "Redis服务"

# 额外清理：根据进程名停止可能残留的进程
Write-Host ""
Write-Host "清理残留进程..." -ForegroundColor Cyan
Stop-ProcessByName -ProcessName "redis-server" -ServiceName "Redis残留进程"
Stop-ProcessByName -ProcessName "node" -ServiceName "Node.js残留进程"

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "           停止完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

Write-Host "所有KamaChat相关服务已停止" -ForegroundColor White
Write-Host ""
Write-Host "按任意键退出..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")