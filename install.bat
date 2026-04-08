@echo off
setlocal enabledelayedexpansion
title Cinelink Monitoring System Installer
color 0F

:menu
cls
echo.
echo  ##########################################################
echo  #                                                        #
echo  #          CINELINK MONITORING SYSTEM DEPLOY             #
echo  #                Build ^& Install Utility                 #
echo  #                                                        #
echo  ##########################################################
echo.
echo   [1] Встановити / Оновити (Install / Update)
echo   [2] Полагодити (Repair - Видалити і поставити з нуля)
echo   [3] Видалити повністю (Uninstall)
echo   [4] Вихід
echo.
set /p choice=" Оберіть дію [1-4]: "

if "%choice%"=="1" goto install
if "%choice%"=="2" goto repair
if "%choice%"=="3" goto uninstall
if "%choice%"=="4" exit
goto menu

:uninstall
echo.
echo  [🗑️] Починаємо повне очищення системи...
taskkill /F /IM master_server.exe /T >nul 2>&1
taskkill /F /IM metrics_agent.exe /T >nul 2>&1

powershell -Command "Unregister-ScheduledTask -TaskName 'ServerMaster' -Confirm:$false -ErrorAction SilentlyContinue"
powershell -Command "Unregister-ScheduledTask -TaskName 'ServerAgent' -Confirm:$false -ErrorAction SilentlyContinue"

powershell -Command "Remove-NetFirewallRule -DisplayName 'Cinelink Master (TCP)' -ErrorAction SilentlyContinue"
powershell -Command "Remove-NetFirewallRule -DisplayName 'Cinelink Agent (TCP)' -ErrorAction SilentlyContinue"
powershell -Command "Remove-NetFirewallRule -DisplayName 'Cinelink WoL (UDP)' -ErrorAction SilentlyContinue"

if exist "master_server.exe" del /f /q "master_server.exe"
if exist "metrics_agent.exe" del /f /q "metrics_agent.exe"
echo  [+] Cinelink успішно видалено з системи!
pause
goto menu

:install
echo.
echo  [1/4] --^> Stopping active processes and cleaning...
taskkill /F /IM master_server.exe /T >nul 2>&1
taskkill /F /IM metrics_agent.exe /T >nul 2>&1

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
goto menu

:repair
echo.
echo  [🔧] Починаємо ремонт (Repair)...
call :uninstall
call :install
goto menu

:error
echo.
echo  !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
echo   ERROR: Compilation failed.
echo  !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
pause
goto menu