@echo off
echo =============================================
echo Auto-BGI 嵌入式全栈应用构建脚本
echo =============================================
echo.

echo.
echo ====== 前端打包开始 ======
cd menu/web
call npm install
call npm run build
call cd ..
call cd ..
echo ====== 前端打包完成 ======


echo ====== 后端打包开始 ======
call go build
echo ====== 后端打包完成 ======
echo ===========后端无窗口打包开始==============
call go build -ldflags="-H=windowsgui" -o auto-bgi无窗口.exe
echo ===========后端无窗口打包完成==============
