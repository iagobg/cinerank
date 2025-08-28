@echo off
REM CineRank Development Setup for Windows

echo.
echo 🎬 CineRank Development Setup
echo =============================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ Go not found. Please install Go 1.25 or later from https://golang.org/dl/
    pause
    exit /b 1
) else (
    for /f "tokens=3" %%a in ('go version') do set GO_VERSION=%%a
    echo ✅ Go found: %GO_VERSION%
)

REM Check if Node.js is installed
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ Node.js not found. Please install Node.js 20+ from https://nodejs.org/
    pause
    exit /b 1
) else (
    for /f %%a in ('node --version') do set NODE_VERSION=%%a
    echo ✅ Node.js found: %NODE_VERSION%
)

REM Check if we're in the right directory
if not exist go.mod (
    echo ❌ go.mod not found. Please run this from the cinerank directory.
    pause
    exit /b 1
)

echo ✅ In project directory
echo.

REM Check DATABASE_URL
if "%DATABASE_URL%"=="" (
    echo ⚠️  DATABASE_URL not set. 
    echo    Please set your Neon database URL:
    echo    set DATABASE_URL=postgresql://username:password@hostname/database?sslmode=require
    echo.
) else (
    echo ✅ DATABASE_URL is configured
)

echo 📦 Installing dependencies...
echo.

REM Install Go dependencies
echo Installing Go tools...
go install github.com/a-h/templ/cmd/templ@latest
go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest

REM Install Node dependencies
echo Installing Node.js dependencies...
npm install

echo.
echo 🎯 Setup complete! Next steps:
echo ==============================
echo.
if "%DATABASE_URL%"=="" (
    echo 1. Set your DATABASE_URL environment variable
    echo 2. Run: make migrate-up
) else (
    echo 1. Run: make migrate-up
)
echo 2. Open 3 separate command prompts and run:
echo    - Terminal 1: templ generate --watch
echo    - Terminal 2: npm run dev:css  
echo    - Terminal 3: make run
echo.
echo 3. Visit http://localhost:8080
echo.
echo Press any key to exit...
pause >nul