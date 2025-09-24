@echo off
echo Testing Docker build locally...

REM Test if Docker is available
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker is not installed or not in PATH
    exit /b 1
)

echo ✅ Docker is available

REM Test Docker build
echo Building Docker image...
docker build -t devhive-test .
if %errorlevel% neq 0 (
    echo ❌ Docker build failed
    exit /b 1
)

echo ✅ Docker build successful

REM Test if the image runs
echo Testing if image runs...
docker run --rm devhive-test echo "Hello from container"
if %errorlevel% neq 0 (
    echo ❌ Docker image failed to run
    exit /b 1
)

echo ✅ Docker image runs successfully
echo 🎉 All Docker tests passed!
