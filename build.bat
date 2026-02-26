@echo off
setlocal EnableDelayedExpansion

echo ====== Auto-BGI构建脚本 ======
echo.

set versionFile=menu\web\version.txt
set indexFile=menu\web\index.html
set backupFile=menu\web\index.html.bak

REM --- 如果没有 version.txt 就创建初始值 0.0.0 ---
if not exist "%versionFile%" (
    echo 0.0.0> "%versionFile%"
    echo 已初始化 %versionFile% 为 0.0.0
)

REM --- 读取旧版本（去除行尾空白） ---
for /f "usebackq tokens=1 delims= " %%v in ("%versionFile%") do set OLD_VERSION=%%v
echo 当前版本号：%OLD_VERSION%

REM --- 解析 major.minor.patch ---
for /f "tokens=1-3 delims=." %%a in ("%OLD_VERSION%") do (
    set MAJOR=%%a
    set MINOR=%%b
    set PATCH=%%c
)

if "%MAJOR%"=="" set MAJOR=0
if "%MINOR%"=="" set MINOR=0
if "%PATCH%"=="" set PATCH=0

REM --- 把 PATCH + 1 ---
set /a PATCH+=1
set NEW_VERSION=%MAJOR%.%MINOR%.%PATCH%
echo 新版本号：%NEW_VERSION%

REM --- 备份 index.html ---
if exist "%indexFile%" (
    copy /Y "%indexFile%" "%backupFile%" >nul
)

REM --- 写回版本号 ---
> "%versionFile%" echo %NEW_VERSION%
powershell -Command "(Get-Content -Raw '%indexFile%') -replace '(?<=autobgi【).*?(?=】)','%NEW_VERSION%' | Out-File -Encoding utf8 '%indexFile%'"

echo 已更新 %indexFile% 中的版本号。
echo.

REM ========== 前端打包 ==========
echo ====== 前端打包开始 ======
pushd menu\web
call npm install
call npm run build
popd
echo ====== 前端打包完成 ======
echo.

REM ========== 后端打包 ==========
echo ====== 后端打包开始 ======

REM --- 1. 关键步骤：清理环境 ---
echo 正在清理旧进程和文件...
taskkill /f /im auto-bgi.exe >nul 2>nul
if exist "auto-bgi.exe" (
    del /f /q "auto-bgi.exe"
    if exist "auto-bgi.exe" (
        echo [错误] 无法删除 auto-bgi.exe，请确保程序已关闭或未被其他工具占用。
        pause
        exit /b
    )
)

REM --- 2. 执行 Go 编译 ---
echo 正在编译 Go 程序...
REM -s -w 减小体积，-H=windowsgui 隐藏窗口
go build -ldflags="-s -w -H=windowsgui" -o "auto-bgi.exe"

REM --- 3. 检查编译是否成功 ---
if %errorlevel% neq 0 (
    echo.
    echo [严重错误] Go 编译失败！请检查上方报错信息。
    pause
    exit /b
)

REM --- 4. 执行 UPX 压缩 ---
where upx >nul 2>nul
if %errorlevel% equ 0 (
    echo.
    echo 正在使用 UPX 压缩 auto-bgi.exe ...
    REM 增加 -f 强制覆盖，确保压缩顺利
    upx -f --best "auto-bgi.exe"
    if %errorlevel% neq 0 (
        echo [警告] UPX 压缩失败，但程序已生成。
    ) else (
        echo UPX 压缩完成！
    )
) else (
    echo.
    echo [警告] 未检测到 UPX 命令。跳过压缩步骤。
)

echo ===========后端无窗口打包完成==============
echo.
echo ====== 构建成功！新版本号：%NEW_VERSION% ======
pause