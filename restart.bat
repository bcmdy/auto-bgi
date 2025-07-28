@echo off
echo 正在重启程序...

rem
taskkill /f /im auto-bgi.exe

timeout /t 1

rem
start "auto-bgi.exe" "auto-bgi.exe"

echo 程序已重启
