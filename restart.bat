@echo off
:: 1. 切到 UTF-8 代码页，避免中文乱码
chcp 65001 > nul
title 重启 auto-bgi 工具

:: ✅ 强制设置工作目录为脚本所在目录
cd /d "%~dp0"

echo 正在重启 auto-bgi.exe ...

:: 1. 尝试结束进程
taskkill /f /im auto-bgi.exe >nul 2>&1

:: 2. 等待确保进程退出（最多等 5 秒）
set "retry=0"
:wait_loop
tasklist /fi "imagename eq auto-bgi.exe" | find /i "auto-bgi.exe" >nul
if %errorlevel%==0 (
    set /a retry+=1
    if %retry% lss 5 (
        timeout /t 1 >nul
        goto :wait_loop
    )
)

:: 3. 启动程序
start "" "%~dp0auto-bgi.exe"

echo auto-bgi.exe 已重新启动，3秒后自动关闭窗口
timeout /t 3 /nobreak >nul
exit