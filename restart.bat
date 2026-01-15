@echo off
chcp 65001 >nul
setlocal EnableExtensions

:: ===== 配置 =====
set "APP=auto-bgi.exe"
set "ARGS=OneLong"
:: 默认工作目录为脚本所在目录（可改为绝对路径）
set "APPDIR=%~dp0"
set "LOG=%APPDIR%restart_auto-bgi.log"

:: 记录开始
echo [INFO] ===== 开始: %DATE% %TIME% =====> "%LOG%"
echo [INFO] 脚本目录: "%APPDIR%" >> "%LOG%"
echo [INFO] 目标程序: "%APP% %ARGS%" >> "%LOG%"

:: 使用 start 启动一个独立的 PowerShell 进程来做“杀进程 + 等待 + 启动”
:: 这样可以避免“不能终止自身”的问题（脱离调用链）
start "" powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
  "$ErrorActionPreference='Continue';" ^
  "$exe = '%APPDIR%%APP%';" ^
  "$args = '%ARGS%';" ^
  "$log = '%LOG%';" ^
  "Add-Content -Path $log -Value ('[PS   ] 开始: ' + (Get-Date));" ^
  "try {" ^
    "Add-Content -Path $log -Value '[PS   ] 正在停止进程: auto-bgi';" ^
    "Get-Process -Name 'auto-bgi' -ErrorAction SilentlyContinue | ForEach-Object { Stop-Process -Id $_.Id -Force -ErrorAction SilentlyContinue };" ^
  " } catch { Add-Content -Path $log -Value ('[PSERR] Stop-Process 出错: ' + $_.Exception.Message) };" ^
  "for($i=0;$i -lt 5;$i++) { if((Get-Process -Name 'auto-bgi' -ErrorAction SilentlyContinue).Count -eq 0) { break } ; Start-Sleep -Seconds 1 };" ^
  "if((Get-Process -Name 'auto-bgi' -ErrorAction SilentlyContinue).Count -ne 0) { Add-Content -Path $log -Value '[PSERR] 进程仍在运行，放弃重启'; exit 1 };" ^
  "try {" ^
    "Add-Content -Path $log -Value ('[PS   ] 启动: ' + $exe + ' ' + $args);" ^
    "Start-Process -FilePath $exe -ArgumentList $args -WorkingDirectory '%APPDIR%' -WindowStyle Normal;" ^
  " } catch { Add-Content -Path $log -Value ('[PSERR] 启动失败: ' + $_.Exception.Message); exit 2 };" ^
  "Start-Sleep -Seconds 2;" ^
  "if((Get-Process -Name 'auto-bgi' -ErrorAction SilentlyContinue).Count -eq 0) { Add-Content -Path $log -Value '[PSERR] 启动后未检测到进程'; exit 3 } else { Add-Content -Path $log -Value '[PSOK ] 已检测到进程'; exit 0 };" ^
  " | Out-File -FilePath '%LOG%' -Append -Encoding UTF8"

echo [INFO] 已委托独立 PowerShell 进程执行“杀+启”，脚本退出（日志：%LOG%）
timeout /t 3 /nobreak >nul
exit /b 0
