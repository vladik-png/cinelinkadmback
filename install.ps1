if (-Not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Host "Requesting Administrator privileges..." -ForegroundColor Yellow
    Start-Process powershell -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`"" -Verb RunAs
    exit
}

$currentDir = $PSScriptRoot
$backendExe = Join-Path -Path $currentDir -ChildPath "backend.exe"
$oldTasks = @("ServerMaster", "ServerAgent", "ServerTerminal", "SystemMetricsAgent", "SystemMasterServer", "CinelinkBackend")

function Show-Menu {
    Clear-Host
    Write-Host "----------------------------------------------------------" -ForegroundColor Gray
    Write-Host "     CINELINK MONITORING SYSTEM DEPLOY" -ForegroundColor Cyan
    Write-Host "                Build & Install Utility" -ForegroundColor DarkCyan
    Write-Host "----------------------------------------------------------" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  [1] Install / Update" -ForegroundColor Green
    Write-Host "  [2] Repair / Reinstall" -ForegroundColor Yellow
    Write-Host "  [3] Uninstall" -ForegroundColor Red
    Write-Host "  [4] Exit" -ForegroundColor White
    Write-Host ""
    $choice = Read-Host " Choose an action [1-4]"
    return $choice
}

function Do-Uninstall {
    Write-Host "`n[] Removing Cinelink components..." -ForegroundColor Yellow
    Stop-Process -Name "backend", "master_server", "metrics_agent", "terminal_proxy" -Force -ErrorAction SilentlyContinue

    foreach ($taskName in $oldTasks) {
        if (Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue) {
            Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
            Write-Host "  [-] Removed task: $taskName" -ForegroundColor DarkGray
        }
    }

    $rules = @("Cinelink Backend (TCP)", "Cinelink Master (TCP)", "Cinelink Agent (TCP)", "Cinelink WoL (UDP)")
    foreach ($rule in $rules) {
        Remove-NetFirewallRule -DisplayName $rule -ErrorAction SilentlyContinue | Out-Null
    }

    if (Test-Path $backendExe) { Remove-Item $backendExe -Force }
    $oldMaster = Join-Path -Path $currentDir -ChildPath "master_server.exe"
    $oldAgent = Join-Path -Path $currentDir -ChildPath "metrics_agent.exe"
    if (Test-Path $oldMaster) { Remove-Item $oldMaster -Force }
    if (Test-Path $oldAgent) { Remove-Item $oldAgent -Force }

    Write-Host "  [+] Cinelink components removed successfully!" -ForegroundColor Green
    Write-Host "Press any key to continue..."
    $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown") | Out-Null
}

function Do-Install {
    Write-Host "`n[1/4] Stopping active processes and cleaning..." -ForegroundColor Yellow
    Stop-Process -Name "backend", "master_server", "metrics_agent" -Force -ErrorAction SilentlyContinue

    foreach ($taskName in $oldTasks) {
        if (Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue) {
            Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
        }
    }

    Write-Host "`n[2/4] Updating Go modules..." -ForegroundColor Yellow
    Push-Location $currentDir
    go mod tidy

    Write-Host "`n[3/4] Compiling fresh binary..." -ForegroundColor Yellow
    Write-Host "      * Building Backend Server..." -ForegroundColor Cyan
    go build -o backend.exe main.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  [!] ERROR: Compilation failed." -ForegroundColor Red
        Pop-Location
        Write-Host "Press any key to continue..."
        $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown") | Out-Null
        return
    }
    Pop-Location

    Write-Host "`n[4/4] Registering System Services & Firewall..." -ForegroundColor Yellow
    if (Test-Path $backendExe) {
        $action = New-ScheduledTaskAction -Execute $backendExe -WorkingDirectory $currentDir
        $trigger = New-ScheduledTaskTrigger -AtStartup
        $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit 0
        Register-ScheduledTask -Action $action -Trigger $trigger -Settings $settings -TaskName "CinelinkBackend" -Description "Cinelink Unified Backend Server" -User "NT AUTHORITY\SYSTEM" -RunLevel Highest -Force | Out-Null
        Start-ScheduledTask -TaskName "CinelinkBackend"
        Write-Host "  [*] Service 'CinelinkBackend' -> ONLINE" -ForegroundColor Green
    }

    Remove-NetFirewallRule -DisplayName "Cinelink Backend (TCP)" -ErrorAction SilentlyContinue | Out-Null
    New-NetFirewallRule -DisplayName "Cinelink Backend (TCP)" -Direction Inbound -LocalPort 8081 -Protocol TCP -Action Allow -Profile Any | Out-Null
    Remove-NetFirewallRule -DisplayName "Cinelink WoL (UDP)" -ErrorAction SilentlyContinue | Out-Null
    New-NetFirewallRule -DisplayName "Cinelink WoL (UDP)" -Direction Inbound -LocalPort 9 -Protocol UDP -Action Allow -Profile Any | Out-Null
    Write-Host "  [+] Firewall rules successfully injected." -ForegroundColor Green

    Write-Host "`nSystem Verification..." -ForegroundColor Yellow
    Start-Sleep -Seconds 3

    $connMaster = Get-NetTCPConnection -LocalPort 8081 -State Listen -ErrorAction SilentlyContinue
    if ($connMaster) {
        Write-Host "  [✔] SUCCESS: Backend Port 8081 is responding." -ForegroundColor Green
        Start-Process "http://127.0.0.1:8081/system-metrics"
    } else {
        Write-Host "  [✘] WARNING: Port 8081 not found active yet." -ForegroundColor Red
    }

    Write-Host "`n==========================================================" -ForegroundColor Gray
    Write-Host " STATUS: DEPLOYMENT FINISHED SUCCESSFULLY" -ForegroundColor Cyan
    Write-Host "==========================================================" -ForegroundColor Gray
    Write-Host "Press any key to continue..."
    $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown") | Out-Null
}

while ($true) {
    $choice = Show-Menu
    switch ($choice) {
        "1" { Do-Install }
        "2" { Do-Uninstall; Do-Install }
        "3" { Do-Uninstall }
        "4" { exit }
    }
}
