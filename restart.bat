@echo off
:: 切换 UTF-8，避免中文乱码（需支持 UTF-8 字体）
chcp 65001 >nul
title 重启 auto-bgi 工具

:: 切换到脚本所在目录
cd /d "%~dp0"

echo 正在关闭 auto-bgi.exe ...

:: 结束进程
taskkill /f /im auto-bgi.exe >nul 2>&1

:: 等待进程完全退出（最多5秒）
setlocal enabledelayedexpansion
set retry=0
:wait_loop
tasklist /fi "imagename eq auto-bgi.exe" | find /i "auto-bgi.exe" >nul
if !errorlevel!==0 (
    set /a retry+=1
    if !retry! lss 5 (
        timeout /t 1 >nul
        goto :wait_loop
    )
)
endlocal

:: 再次确认是否还在运行
tasklist /fi "imagename eq auto-bgi.exe" | find /i "auto-bgi.exe" >nul
if %errorlevel%==0 (
    echo 错误：auto-bgi.exe 无法关闭，重启失败！
    timeout /t 5 >nul
    exit /b 1
)

echo auto-bgi.exe 已关闭，正在启动...
start "" /d "%~dp0" "auto-bgi.exe"

echo auto-bgi.exe 已重新启动，3秒后关闭窗口
timeout /t 3 /nobreak >nul
exit
