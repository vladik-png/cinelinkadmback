$currentDir = $PSScriptRoot
$masterExe = Join-Path -Path $currentDir -ChildPath "master_server.exe"
$agentExe = Join-Path -Path $currentDir -ChildPath "metrics_agent.exe"
$oldTasks = @("ServerMaster", "ServerAgent", "SystemMetricsAgent", "SystemMasterServer")

Clear-Host
Write-Host "----------------------------------------------------------" -ForegroundColor Gray
Write-Host "    CINELINK INFRASTRUCTURE INSTALLER v1.0" -ForegroundColor Cyan
Write-Host "----------------------------------------------------------" -ForegroundColor Gray

Write-Host "`n[1/3] Cleaning environment..." -ForegroundColor Yellow
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
Stop-Process -Name "master_server", "metrics_agent" -Force -ErrorAction SilentlyContinue
Write-Host "  [+] Environment is clean." -ForegroundColor Green


Write-Host "`n[2/3] Registering System Services..." -ForegroundColor Yellow
for ($i = 1; $i -le 100; $i+=10) {
    Write-Progress -Activity "Deploying Cinelink" -Status "Registering binaries: $i%" -PercentComplete $i
    Start-Sleep -Milliseconds 50
}

function Deploy-Task($name, $path, $desc) {
    if (Test-Path $path) {
        $action = New-ScheduledTaskAction -Execute $path -WorkingDirectory $currentDir
        $trigger = New-ScheduledTaskTrigger -AtStartup
        $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit 0
        Register-ScheduledTask -Action $action -Trigger $trigger -Settings $settings -TaskName $name -Description $desc -User "NT AUTHORITY\SYSTEM" -RunLevel Highest -Force
        Start-ScheduledTask -TaskName $name
        Write-Host "  [*] Service '$name' -> ONLINE" -ForegroundColor Green
    }
}

Deploy-Task "ServerMaster" $masterExe "Cinelink Master Monitoring Server"
Deploy-Task "ServerAgent" $agentExe "Cinelink Metrics Agent"

Write-Host "`n[3/3] System Verification..." -ForegroundColor Yellow
Write-Progress -Activity "Deploying Cinelink" -Status "Final checks: 100%" -PercentComplete 100
Start-Sleep -Seconds 4

$conn = Get-NetTCPConnection -LocalPort 8082 -State Listen -ErrorAction SilentlyContinue
if ($conn) {
    Write-Host "  [✔] SUCCESS: Port 8082 is responding." -ForegroundColor Green
    Write-Host "  [✔] Launching monitoring dashboard..." -ForegroundColor Cyan
    Start-Process "http://127.0.0.1:8082/system-metrics"
} else {
    Write-Host "  [✘] ERROR: Port 8082 not found. Check server logs." -ForegroundColor Red
}

Write-Host "`n----------------------------------------------------------" -ForegroundColor Gray
Write-Host "   DEPLOYMENT COMPLETE! Closing in 10s." -ForegroundColor Cyan
Write-Host "----------------------------------------------------------" -ForegroundColor Gray
Start-Sleep -Seconds 10