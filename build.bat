@echo off
echo =============================================
echo Auto-BGI 嵌入式全栈应用构建脚本
echo =============================================
echo.

REM 检查是否安装了npm
where npm >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 未找到npm，请确保已安装Node.js
    pause
    exit /b 1
)

REM 检查是否安装了go
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 未找到go，请确保已安装Go语言环境
    pause
    exit /b 1
)

echo.
echo ====== 前端打包开始 ======
cd web
if exist node_modules (
    echo 已存在 node_modules，跳过安装依赖
) else (
    echo 安装前端依赖...
    npm install
)
npm run build
if errorlevel 1 (
    echo ❌ 前端打包失败！
    pause
    exit /b 1
)
cd ..
echo ====== 前端打包完成 ======

echo.
echo 当前目录：
cd

echo ====== 后端打包开始 ======
go build -o AutoBGI.exe main.go
if errorlevel 1 (
    echo ❌ 后端打包失败！
    pause
    exit /b 1
)
echo ====== 后端打包完成，生成文件 AutoBGI.exe ======

pause
