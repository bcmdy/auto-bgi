@echo off
echo =============================================
echo Auto-BGI 嵌入式全栈应用构建脚本
echo =============================================
echo.

echo.
echo ====== 前端打包开始 ======
cd web
call npm install
call npm run build
call cd ..
echo ====== 前端打包完成 ======


echo ====== 后端打包开始 ======
call go build
echo ====== 后端打包完成 ======
