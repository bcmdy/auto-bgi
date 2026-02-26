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

REM --- 解析 major.minor.patch（如果缺少部分，补 0） ---
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

REM --- 组装新版本号 ---
set NEW_VERSION=%MAJOR%.%MINOR%.%PATCH%
echo 新版本号：%NEW_VERSION%

REM --- 备份 index.html（以防万一） ---
if exist "%indexFile%" (
    copy /Y "%indexFile%" "%backupFile%" >nul
    echo 已备份 %indexFile% -> %backupFile%
)

REM --- 写回 version.txt ---
> "%versionFile%" echo %NEW_VERSION%

REM --- 使用 PowerShell 安全替换 index.html 中 autobgi【...】 的部分，写回 UTF-8 ---
powershell -Command "(Get-Content -Raw '%indexFile%') -replace '(?<=autobgi【).*?(?=】)','%NEW_VERSION%' | Out-File -Encoding utf8 '%indexFile%'"

echo 已更新 %indexFile% 中的版本号为：%NEW_VERSION%
echo.

REM ========== 前端打包 ==========
pushd menu\web
call npm install
call npm run build
popd
echo ====== 前端打包完成 ======


echo ====== 后端打包开始 ======

REM --- 1. 执行 Go 编译 ---
REM 注意：这里我保留了你的 -ldflags="-H=windowsgui" 参数（隐藏控制台窗口）
REM 同时也建议加上 -s -w 来去除调试信息，这本身就能减小体积
echo 正在编译 Go 程序...
call go build -ldflags="-s -w -H=windowsgui" -o "auto-bgi.exe"

REM --- 2. 执行 UPX 压缩 ---
REM 检查 upx 是否存在于环境变量中，或者你可以指定 upx.exe 的绝对路径
where upx >nul 2>nul
if %errorlevel% equ 0 (
    echo.
    echo 正在使用 UPX 压缩 auto-bgi.exe ...
    REM --best 表示最高压缩率，-f 表示强制覆盖（如果有必要）
    upx --best "auto-bgi.exe"
    echo UPX 压缩完成！
) else (
    echo.
    echo [警告] 未检测到 UPX 命令。跳过压缩步骤。
    echo 请确保 upx.exe 在你的系统 PATH 环境变量中，或放在当前目录下。
)

echo ===========后端无窗口打包完成==============

echo.
echo ====== 构建完成，新版本号：%NEW_VERSION% ======
