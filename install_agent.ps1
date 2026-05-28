$currentDir = $PSScriptRoot
$masterExe = Join-Path -Path $currentDir -ChildPath "master_server.exe"
$agentExe = Join-Path -Path $currentDir -ChildPath "metrics_agent.exe"
$oldTasks = @("ServerMaster", "ServerAgent", "ServerTerminal", "SystemMetricsAgent", "SystemMasterServer")

Clear-Host
Write-Host "----------------------------------------------------------" -ForegroundColor Gray
Write-Host "     CINELINK INFRASTRUCTURE INSTALLER v1.1" -ForegroundColor Cyan
Write-Host "----------------------------------------------------------" -ForegroundColor Gray

Write-Host "`n[1/4] Cleaning environment..." -ForegroundColor Yellow
for ($i = 1; $i -le 100; $i+=20) {
    Write-Progress -Activity "Deploying Cinelink" -Status "Cleaning old services: $i%" -PercentComplete $i
    Start-Sleep -Milliseconds 100
}

foreach ($taskName in $oldTasks) {
    if (Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue) {
        Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
        Write-Host "  [-] Removed task: $taskName" -ForegroundColor DarkGray
    }
}
Stop-Process -Name "master_server", "metrics_agent", "terminal_proxy" -Force -ErrorAction SilentlyContinue
Write-Host "  [+] Environment is clean." -ForegroundColor Green


Write-Host "`n[2/4] Registering System Services..." -ForegroundColor Yellow
for ($i = 1; $i -le 100; $i+=10) {
    Write-Progress -Activity "Deploying Cinelink" -Status "Registering binaries: $i%" -PercentComplete $i
    Start-Sleep -Milliseconds 50
}

function Register-Task($name, $path, $desc) {
    if (Test-Path $path) {
        $action = New-ScheduledTaskAction -Execute $path -WorkingDirectory $currentDir
        $trigger = New-ScheduledTaskTrigger -AtStartup
        $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit 0
        Register-ScheduledTask -Action $action -Trigger $trigger -Settings $settings -TaskName $name -Description $desc -User "NT AUTHORITY\SYSTEM" -RunLevel Highest -Force | Out-Null
        Start-ScheduledTask -TaskName $name
        Write-Host "  [*] Service '$name' -> ONLINE" -ForegroundColor Green
    } else {
        Write-Host "  [!] WARNING: Binary '$path' not found! Skipping..." -ForegroundColor Magenta
    }
}

Register-Task "ServerMaster" $masterExe "Cinelink Master Monitoring Server"
Register-Task "ServerAgent" $agentExe "Cinelink Metrics Agent"

Write-Host "`n[3/4] Configuring Windows Firewall..." -ForegroundColor Yellow
for ($i = 1; $i -le 100; $i+=25) {
    Write-Progress -Activity "Deploying Cinelink" -Status "Adding firewall rules: $i%" -PercentComplete $i
    Start-Sleep -Milliseconds 50
}

function Enable-Port($name, $port, $proto) {
    Remove-NetFirewallRule -DisplayName $name -ErrorAction SilentlyContinue | Out-Null
    New-NetFirewallRule -DisplayName $name -Direction Inbound -LocalPort $port -Protocol $proto -Action Allow -Profile Any | Out-Null
    Write-Host "  [+] Opened port: $port ($proto)" -ForegroundColor DarkGray
}

Enable-Port "Cinelink Master (TCP)" 8081 "TCP"
Enable-Port "Cinelink Agent (TCP)" 8082 "TCP"
Enable-Port "Cinelink WoL (UDP)" 9 "UDP"
Write-Host "  [+] Firewall rules successfully injected." -ForegroundColor Green


Write-Host "`n[4/4] System Verification..." -ForegroundColor Yellow
Write-Progress -Activity "Deploying Cinelink" -Status "Final checks: 100%" -PercentComplete 100
Start-Sleep -Seconds 3

$connMaster = Get-NetTCPConnection -LocalPort 8081 -State Listen -ErrorAction SilentlyContinue
if ($connMaster) {
    Write-Host "  [✔] SUCCESS: Master Port 8081 is responding." -ForegroundColor Green
    Write-Host "  [✔] Launching monitoring dashboard..." -ForegroundColor Cyan
    Start-Process "http://127.0.0.1:8081/system-metrics"
} else {
    Write-Host "  [✘] WARNING: Port 8081 not found active yet. Check server logs if it fails." -ForegroundColor Yellow
}

Write-Host "`n----------------------------------------------------------" -ForegroundColor Gray
Write-Host "   DEPLOYMENT COMPLETE! Closing in 10s." -ForegroundColor Cyan
Write-Host "----------------------------------------------------------" -ForegroundColor Gray
Start-Sleep -Seconds 10