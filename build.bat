@echo off
echo =============================================
echo Auto-BGI Ƕ��ʽȫջӦ�ù����ű�
echo =============================================
echo.

echo.
echo ====== ǰ�˴����ʼ ======
cd web
call npm install
call npm run build
call cd ..
echo ====== ǰ�˴����� ======


echo ====== ��˴����ʼ ======
call go build
echo ====== ��˴����� ======
