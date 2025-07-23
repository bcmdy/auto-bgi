@echo off
echo 正在重启程序...

rem 强制杀死程序，改成你的程序名
taskkill /f /im auto-bgi.exe

timeout /t 1

rem 启动程序（写成绝对路径或相对路径均可）
start "" "auto-bgi.exe"

echo 程序已重启
