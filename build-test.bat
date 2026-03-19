@echo off
setlocal EnableDelayedExpansion
chcp 65001 >nul

echo ====== Auto-BGI Build Script ======
echo.

set versionFile=menu\web\version.txt
set indexFile=menu\web\index.html
set backupFile=menu\web\index.html.bak

REM --- Initialize version.txt to 0.0.0 if missing ---
if not exist "%versionFile%" (
    echo 0.0.0> "%versionFile%"
    echo Initialized %versionFile% to 0.0.0
)

REM --- Read old version ---
for /f "usebackq tokens=1 delims= " %%v in ("%versionFile%") do set OLD_VERSION=%%v
echo Current Version: %OLD_VERSION%

REM --- Parse major.minor.patch ---
for /f "tokens=1-3 delims=." %%a in ("%OLD_VERSION%") do (
    set MAJOR=%%a
    set MINOR=%%b
    set PATCH=%%c
)

if "%MAJOR%"=="" set MAJOR=0
if "%MINOR%"=="" set MINOR=0
if "%PATCH%"=="" set PATCH=0

REM --- Increment PATCH ---
set /a PATCH+=1

REM --- Assemble new version ---
set NEW_VERSION=%MAJOR%.%MINOR%.%PATCH%
echo New Version: %NEW_VERSION%

REM --- Backup index.html ---
if exist "%indexFile%" (
    copy /Y "%indexFile%" "%backupFile%" >nul
    echo Backed up %indexFile% -> %backupFile%
)

REM --- Update version.txt ---
> "%versionFile%" echo %NEW_VERSION%

REM --- Update index.html using PowerShell (UTF-8 safe) ---
REM Uses [System.IO.File] to read/write UTF-8 without BOM to avoid issues
powershell -NoProfile -Command "$content = [System.IO.File]::ReadAllText('%indexFile%', [System.Text.Encoding]::UTF8); $newContent = $content -replace '(?<=autobgi\u3010).*?(?=\u3011)','%NEW_VERSION%'; [System.IO.File]::WriteAllText('%indexFile%', $newContent, [System.Text.Encoding]::UTF8)"

echo Updated version in %indexFile% to %NEW_VERSION%
echo.

REM ========== Frontend Build ==========
echo Building Frontend...
pushd menu\web
call npm install
call npm run build
popd
echo Frontend Build Complete

echo.
echo ====== Backend Build ======

REM --- 1. Go Build ---
echo Building Go binary...
call go build -o "auto-bgi.exe"

REM --- 2. UPX Compression ---
where upx >nul 2>nul
if %errorlevel% equ 0 (
    echo.
    echo Compressing with UPX...
    upx --best "auto-bgi.exe"
    echo UPX Compression Complete
) else (
    echo.
    echo [Warning] UPX not found. Skipping compression.
)

echo.
echo ====== Build Complete! New Version: %NEW_VERSION% ======
