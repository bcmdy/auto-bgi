@echo off
:: 关闭命令回显
title 重启 auto-bgi 工具

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

:: 3. 启动程序（第一个引号是窗口标题，可为空）
start "" "%~dp0auto-bgi.exe"

echo auto-bgi.exe 已重新启动
pause