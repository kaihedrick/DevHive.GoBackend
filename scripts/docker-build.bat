@echo off
REM DevHive Backend Docker Build Script for Windows
REM This script helps build, run, and manage Docker containers

setlocal enabledelayedexpansion

REM Default values
set ENVIRONMENT=dev
set IMAGE_NAME=devhive-backend
set TAG=latest
set COMPOSE_FILE=docker-compose.yml

REM Function to print colored output
:print_status
echo [INFO] %~1
goto :eof

:print_success
echo [SUCCESS] %~1
goto :eof

:print_warning
echo [WARNING] %~1
goto :eof

:print_error
echo [ERROR] %~1
goto :eof

REM Function to show usage
:show_usage
echo Usage: %~nx0 [COMMAND] [OPTIONS]
echo.
echo Commands:
echo   build       Build the Docker image
echo   run         Run the application (dev mode)
echo   run-prod    Run the application (production mode)
echo   stop        Stop all containers
echo   clean       Clean up containers and images
echo   logs        Show container logs
echo   shell       Access container shell
echo   test        Run health checks
echo.
echo Options:
echo   -t TAG        Docker image tag (default: latest)
echo   -e ENV        Environment (dev/prod, default: dev)
echo   -h            Show this help message
echo.
echo Examples:
echo   %~nx0 build                    # Build image with latest tag
echo   %~nx0 run                      # Run in development mode
echo   %~nx0 run-prod                 # Run in production mode
echo   %~nx0 stop                     # Stop all containers
echo   %~nx0 test                     # Test all endpoints
goto :eof

REM Function to build Docker image
:build_image
call :print_status "Building Docker image: %IMAGE_NAME%:%TAG%"

REM Check if .env file exists
if not exist .env (
    call :print_warning ".env file not found. Creating from env.example..."
    if exist env.example (
        copy env.example .env >nul
        call :print_warning "Please update .env file with your configuration"
    ) else (
        call :print_error "env.example file not found. Please create .env file manually."
        exit /b 1
    )
)

docker build -t %IMAGE_NAME%:%TAG% .
if %ERRORLEVEL% EQU 0 (
    call :print_success "Docker image built successfully: %IMAGE_NAME%:%TAG%"
) else (
    call :print_error "Failed to build Docker image"
    exit /b 1
)
goto :eof

REM Function to run development environment
:run_dev
call :print_status "Starting development environment..."

if not exist .env (
    call :print_error ".env file not found. Please create it first."
    exit /b 1
)

docker-compose -f %COMPOSE_FILE% up -d
if %ERRORLEVEL% EQU 0 (
    call :print_success "Development environment started"
    call :print_status "Application available at: http://localhost:8080"
    call :print_status "gRPC server at: localhost:8081"
    call :print_status "Health check at: http://localhost:8080/health"
) else (
    call :print_error "Failed to start development environment"
    exit /b 1
)
goto :eof

REM Function to run production environment
:run_prod
call :print_status "Starting production environment..."

if not exist .env (
    call :print_error ".env file not found. Please create it first."
    exit /b 1
)

set COMPOSE_FILE=docker-compose.prod.yml
docker-compose -f %COMPOSE_FILE% up -d
if %ERRORLEVEL% EQU 0 (
    call :print_success "Production environment started"
    call :print_status "Application available at: http://localhost:8080"
    call :print_status "Health check at: http://localhost:8080/health"
) else (
    call :print_error "Failed to start production environment"
    exit /b 1
)
goto :eof

REM Function to stop containers
:stop_containers
call :print_status "Stopping all containers..."
docker-compose -f %COMPOSE_FILE% down
docker-compose -f docker-compose.prod.yml down 2>nul
call :print_success "All containers stopped"
goto :eof

REM Function to clean up
:clean_up
call :print_status "Cleaning up containers and images..."
docker-compose -f %COMPOSE_FILE% down -v --rmi all
docker-compose -f docker-compose.prod.yml down -v --rmi all 2>nul
docker system prune -f
call :print_success "Cleanup completed"
goto :eof

REM Function to show logs
:show_logs
call :print_status "Showing container logs..."
docker-compose -f %COMPOSE_FILE% logs -f
goto :eof

REM Function to access container shell
:access_shell
call :print_status "Accessing container shell..."
docker exec -it devhive-backend sh
goto :eof

REM Function to test endpoints
:test_endpoints
call :print_status "Testing all endpoints..."

REM Wait for service to be ready
call :print_status "Waiting for service to be ready..."
timeout /t 10 /nobreak >nul

REM Test health endpoint
call :print_status "Testing health endpoint..."
curl -f http://localhost:8080/health >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    call :print_success "Health endpoint: OK"
) else (
    call :print_error "Health endpoint: FAILED"
    exit /b 1
)

REM Test gRPC server
call :print_status "Testing gRPC server..."
netstat -an | findstr :8081 >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    call :print_success "gRPC server: OK"
) else (
    call :print_warning "gRPC server: FAILED (may not be critical)"
)

call :print_success "All endpoint tests completed"
goto :eof

REM Parse command line arguments
:parse_args
if "%~1"=="" goto :execute_command
if "%~1"=="-t" (
    set TAG=%~2
    shift
    shift
    goto :parse_args
)
if "%~1"=="-e" (
    set ENVIRONMENT=%~2
    shift
    shift
    goto :parse_args
)
if "%~1"=="-h" (
    call :show_usage
    exit /b 0
)
if "%~1"=="build" (
    set COMMAND=build
    shift
    goto :parse_args
)
if "%~1"=="run" (
    set COMMAND=run
    shift
    goto :parse_args
)
if "%~1"=="run-prod" (
    set COMMAND=run-prod
    shift
    goto :parse_args
)
if "%~1"=="stop" (
    set COMMAND=stop
    shift
    goto :parse_args
)
if "%~1"=="clean" (
    set COMMAND=clean
    shift
    goto :parse_args
)
if "%~1"=="logs" (
    set COMMAND=logs
    shift
    goto :parse_args
)
if "%~1"=="shell" (
    set COMMAND=shell
    shift
    goto :parse_args
)
if "%~1"=="test" (
    set COMMAND=test
    shift
    goto :parse_args
)
call :print_error "Unknown option: %~1"
call :show_usage
exit /b 1

REM Execute command
:execute_command
if "%COMMAND%"=="" (
    call :print_error "No command specified"
    call :show_usage
    exit /b 1
)

if "%COMMAND%"=="build" (
    call :build_image
) else if "%COMMAND%"=="run" (
    call :run_dev
) else if "%COMMAND%"=="run-prod" (
    call :run_prod
) else if "%COMMAND%"=="stop" (
    call :stop_containers
) else if "%COMMAND%"=="clean" (
    call :clean_up
) else if "%COMMAND%"=="logs" (
    call :show_logs
) else if "%COMMAND%"=="shell" (
    call :access_shell
) else if "%COMMAND%"=="test" (
    call :test_endpoints
) else (
    call :print_error "Unknown command: %COMMAND%"
    exit /b 1
)

exit /b 0
