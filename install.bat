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

echo  [1/3] --^> Updating Go modules...
go mod tidy

echo  [2/3] --^> Compiling binaries for Windows...
echo        * Building Master Server...
go build -o master_server.exe main.go
if %errorlevel% neq 0 goto error

echo        * Building Metrics Agent...
go build -ldflags "-H=windowsgui" -o metrics_agent.exe ./agent/metrix.go
if %errorlevel% neq 0 goto error

echo.
echo  [3/3] --^> Elevating privileges for Installation...
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
echo   ERROR: Compilation failed. Please check your Go code.
echo  !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
pause
exit