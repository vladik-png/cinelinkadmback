$currentDir = $PSScriptRoot
$masterExe = Join-Path -Path $currentDir -ChildPath "master_server.exe"
$agentExe = Join-Path -Path $currentDir -ChildPath "metrics_agent.exe"

$oldTasks = @("ServerMaster", "ServerAgent", "SystemMetricsAgent", "SystemMasterServer")

Write-Host "--- Installing Cinelink Monitoring ---" -ForegroundColor Yellow

foreach ($taskName in $oldTasks) {
    if (Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue) {
        Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
    }
}
Stop-Process -Name "master_server", "metrics_agent" -Force -ErrorAction SilentlyContinue

function Register-And-Start($name, $path) {
    if (Test-Path $path) {
        $action = New-ScheduledTaskAction -Execute $path -WorkingDirectory $currentDir
        $trigger = New-ScheduledTaskTrigger -AtStartup
        $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit 0
        
        Register-ScheduledTask -Action $action -Trigger $trigger -Settings $settings -TaskName $name -User "NT AUTHORITY\SYSTEM" -RunLevel Highest -Force
        Start-ScheduledTask -TaskName $name
        Write-Host "$name Started!" -ForegroundColor Green
    }
}

Register-And-Start -name "ServerMaster" -path $masterExe
Register-And-Start -name "ServerAgent" -path $agentExe

Write-Host "`n Checking local port..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

if (Get-NetTCPConnection -LocalPort 8082 -State Listen -ErrorAction SilentlyContinue) {
    Write-Host " Port 8082 is active! Opening browser..." -ForegroundColor Cyan
    Start-Process "http://127.0.0.1:8082/system-metrics"
} else {
    Write-Host "Master failed to start port 8082." -ForegroundColor Red
}

Write-Host "`nReady! Console will close in 10 seconds..."
Start-Sleep -Seconds 10