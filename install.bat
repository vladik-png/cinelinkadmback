@echo off
setlocal enabledelayedexpansion
title Cinelink Monitoring System Installer

:header
cls
echo.
echo  ##########################################################
echo  #                                                        #
echo  #          CINELINK MONITORING SYSTEM DEPLOY             #
echo  #                Build ^& Install Utility                 #
echo  #                                                        #
echo  ##########################################################
echo.

echo  [1/4] --^> Stopping active processes and cleaning...

taskkill /F /IM master_server.exe /T >nul 2>&1
taskkill /F /IM metrics_agent.exe /T >nul 2>&1

timeout /t 1 /nobreak >nul

if exist "master_server.exe" (
    del /f /q "master_server.exe"
    if exist "master_server.exe" (
        echo  [!] WARNING: Could not delete master_server.exe. Try running as Admin.
    ) else (
        echo        * Old master_server.exe deleted.
    )
)

if exist "metrics_agent.exe" (
    del /f /q "metrics_agent.exe"
    if exist "metrics_agent.exe" (
        echo  [!] WARNING: Could not delete metrics_agent.exe.
    ) else (
        echo        * Old metrics_agent.exe deleted.
    )
)

echo.
echo  [2/4] --^> Updating Go modules...
go mod tidy

echo.
echo  [3/4] --^> Compiling fresh binaries...
echo        * Building Master Server...
go build -o master_server.exe .
if %errorlevel% neq 0 goto error

echo        * Building Metrics Agent...
go build -ldflags "-H=windowsgui" -o metrics_agent.exe ./agent/metrix.go
if %errorlevel% neq 0 goto error

echo.
echo  [4/4] --^> Elevating privileges for Installation...
powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "& {Start-Process powershell.exe -ArgumentList '-NoProfile -ExecutionPolicy Bypass -File ""%~dp0install_agent.ps1""' -Verb RunAs -Wait}"

echo.
echo  ==========================================================
echo   STATUS: DEPLOYMENT FINISHED SUCCESSFULLY
echo  ==========================================================
pause
exit

:error
echo.
echo  !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
echo   ERROR: Compilation failed.
echo  !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
pause
exit