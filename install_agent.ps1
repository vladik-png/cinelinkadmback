$currentDir = $PSScriptRoot

$masterExe = Join-Path -Path $currentDir -ChildPath "master_server.exe"
$agentExe = Join-Path -Path $currentDir -ChildPath "metrics_agent.exe"

$oldTasks = @("ServerMaster", "ServerAgent", "SystemMetricsAgent", "SystemMasterServer")

Write-Host "--- I'm starting installation of system monitoring ---`n" -ForegroundColor Yellow

Write-Host " Cleaning system from old versions..." -ForegroundColor Gray
foreach ($taskName in $oldTasks) {
    if (Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue) {
        Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
        Write-Host "  - Old task '$taskName' removed." -ForegroundColor DarkGray
    }
}

Stop-Process -Name "master_server", "metrics_agent" -Force -ErrorAction SilentlyContinue

function Register-And-Start-Task($name, $path, $description) {
    if (Test-Path $path) {
        $action = New-ScheduledTaskAction -Execute $path -WorkingDirectory $currentDir
        $trigger = New-ScheduledTaskTrigger -AtStartup
        
        $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit 0
        
        Register-ScheduledTask -Action $action -Trigger $trigger -Settings $settings -TaskName $name -Description $description -User "NT AUTHORITY\SYSTEM" -RunLevel Highest -Force
        
        Write-Host "Task '$name' registered successfully!" -ForegroundColor Green
        
        Start-ScheduledTask -TaskName $name
        Write-Host "Task '$name' started!" -ForegroundColor Cyan
    } else {
        Write-Host "Error: File not found at path $path" -ForegroundColor Red
    }
}

Register-And-Start-Task -name "ServerMaster" -path $masterExe -description "Master Monitoring Server (Cinelink)"
Register-And-Start-Task -name "ServerAgent" -path $agentExe -description "Metrics Agent (Cinelink)"

Write-Host "`nInstallation completed! Check port 8082." -ForegroundColor Cyan
Start-Sleep -Seconds 5
netstat -ano | findstr :8082