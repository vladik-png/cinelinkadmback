$currentDir = $PSScriptRoot

$masterExe = Join-Path -Path $currentDir -ChildPath "master_server.exe"
$agentExe = Join-Path -Path $currentDir -ChildPath "metrics_agent.exe"

function Register-MyTask($name, $path, $description) {
    if (Test-Path $path) {
        $action = New-ScheduledTaskAction -Execute $path -WorkingDirectory $currentDir
        $trigger = New-ScheduledTaskTrigger -AtStartup
        $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit 0
        
        Register-ScheduledTask -Action $action -Trigger $trigger -Settings $settings -TaskName $name -Description $description -User "NT AUTHORITY\SYSTEM" -RunLevel Highest -Force
        Write-Host "Task '$name' added successfully!" -ForegroundColor Green
    } else {
        Write-Host "❌ Error: File $path not found!" -ForegroundColor Red
    }
}

Register-MyTask -name "ServerMaster" -path $masterExe -description "Головний сервер моніторингу"
Register-MyTask -name "ServerAgent" -path $agentExe -description "Агент збору метрик"

Write-Host "`n🚀 Installation completed! Programs will start after reboot or by running Start-ScheduledTask." -ForegroundColor Cyan