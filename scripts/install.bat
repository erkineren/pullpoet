@echo off
REM PullPoet Windows Install Script Launcher
REM Downloads and runs the PowerShell installation script

setlocal enabledelayedexpansion

echo 🚀 PullPoet Windows Install Script
echo =====================================
echo.

REM Check if PowerShell is available
powershell -Command "exit" >nul 2>&1
if errorlevel 1 (
    echo [ERROR] PowerShell is not available. Please install PowerShell 5.0 or higher.
    pause
    exit /b 1
)

REM Check PowerShell version
for /f "tokens=*" %%i in ('powershell -Command "$PSVersionTable.PSVersion.Major"') do set PS_VERSION=%%i
if %PS_VERSION% LSS 5 (
    echo [ERROR] PowerShell 5.0 or higher is required. Current version: %PS_VERSION%
    pause
    exit /b 1
)

echo [INFO] PowerShell version: %PS_VERSION%
echo [INFO] Downloading and running installation script...
echo.

REM Download and run the PowerShell script
powershell -Command "& {Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1' -UseBasicParsing | Invoke-Expression} %*"

if errorlevel 1 (
    echo.
    echo [ERROR] Installation failed. Please check the error messages above.
    pause
    exit /b 1
) else (
    echo.
    echo [SUCCESS] Installation completed successfully!
    echo.
    echo Quick start:
    echo   pullpoet --help                    # Show help
    echo   pullpoet --version                 # Show version
    echo.
    echo Example usage:
    echo   set PULLPOET_PROVIDER=openai
    echo   set PULLPOET_MODEL=gpt-3.5-turbo
    echo   set PULLPOET_API_KEY=your-api-key
    echo   pullpoet --target main
    echo.
    echo Learn more: https://github.com/erkineren/pullpoet
    echo.
    pause
) 