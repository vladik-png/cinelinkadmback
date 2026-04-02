@echo off
setlocal enabledelayedexpansion

echo Starting compilation for Windows...

go mod tidy

go build -o master_server.exe main.go
if %errorlevel% neq 0 (
    echo  Error building main.go
    pause
    exit /b
)

go build -ldflags "-H=windowsgui" -o metrics_agent.exe ./agent/metrix.go
if %errorlevel% neq 0 (
    echo  Error compiling agent/metrix.go
    pause
    exit /b
)

echo  Compilation completed successfully!
echo  Starting installation...

powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "& {Start-Process powershell.exe -ArgumentList '-NoProfile -ExecutionPolicy Bypass -File ""%~dp0install_full.ps1""' -Verb RunAs}"

exit