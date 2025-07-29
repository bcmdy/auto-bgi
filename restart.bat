@echo off
:: 关闭命令回显，避免执行过程中显示命令本身
echo 正在重启程序...

:: 强制终止 auto-bgi.exe 进程 (/f 强制终止, /im 指定进程名)
taskkill /f /im auto-bgi.exe

:: 等待1秒，确保进程有足够时间完全退出
timeout /t 3

:: 启动 auto-bgi.exe 程序
start "auto-bgi.exe"

echo 程序已重启
:: 输出重启完成提示信息