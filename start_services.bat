@echo off
chcp 65001 >nul
echo ========================================
echo       KamaChat 服务启动脚本
echo ========================================
echo.

:: 设置颜色
color 0A

:: 检查Redis是否已在运行
echo [1/4] 检查Redis服务状态...
netstat -an | findstr ":6379" >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ Redis服务已在运行
) else (
    echo [2/4] 启动Redis服务...
    if exist "C:\Redis\redis-server.exe" (
        echo 正在启动Redis...
        start "Redis Server" cmd /c "C:\Redis\redis-server.exe C:\Redis\redis.windows.conf"
        timeout /t 3 /nobreak >nul
        echo ✓ Redis服务启动完成
    ) else (
        echo ✗ 错误: 未找到Redis安装路径 C:\Redis\redis-server.exe
        echo   请确保Redis已正确安装在 C:\Redis\ 目录
        pause
        exit /b 1
    )
)

echo.
echo [3/4] 启动Go后端服务...
echo 正在启动后端服务 (端口: 8000)...
start "KamaChat Backend" cmd /k "cd /d %~dp0 && go run cmd/kama_chat_server/main.go"

echo.
echo [4/4] 启动Vue前端服务...
echo 正在启动前端服务 (端口: 8080)...
start "KamaChat Frontend" cmd /k "cd /d %~dp0web\chat-server && npm run serve"

echo.
echo ========================================
echo           启动完成！
echo ========================================
echo.
echo 服务访问地址:
echo   前端应用: http://localhost:8080
echo   后端API:  http://localhost:8000
echo   WebSocket: ws://localhost:8000/wss
echo.
echo 注意: 
echo - 请等待几秒钟让所有服务完全启动
echo - 如需停止服务，请关闭对应的命令行窗口
echo.
echo 按任意键退出...
pause >nul