$currentDir = $PSScriptRoot

$masterExe = Join-Path -Path $currentDir -ChildPath "master_server.exe"
$agentExe = Join-Path -Path $currentDir -ChildPath "metrics_agent.exe"

function Register-And-Start-Task($name, $path, $description) {
    if (Test-Path $path) {
        $action = New-ScheduledTaskAction -Execute $path -WorkingDirectory $currentDir
        $trigger = New-ScheduledTaskTrigger -AtStartup
        
        $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit 0
        
        Register-ScheduledTask -Action $action -Trigger $trigger -Settings $settings -TaskName $name -Description $description -User "NT AUTHORITY\SYSTEM" -RunLevel Highest -Force
        
        Write-Host "Task '$name' added successfully!" -ForegroundColor Green
        
        Start-ScheduledTask -TaskName $name
        Write-Host "Task '$name' started!" -ForegroundColor Cyan
    } else {
        Write-Host "Error: File $path not found at $path" -ForegroundColor Red
    }
}

Write-Host "--- Installing Monitoring System ---`n" -ForegroundColor Yellow

Register-And-Start-Task -name "ServerMaster" -path $masterExe -description "Master Monitoring Server (Cinelink)"
Register-And-Start-Task -name "ServerAgent" -path $agentExe -description "Metrics Agent (Cinelink)"

Write-Host "`n Installation completed. Check port 8082 via netstat." -ForegroundColor Cyan