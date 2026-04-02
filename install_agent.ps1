$exePath = Join-Path -Path $PSScriptRoot -ChildPath "metrics_agent.exe"
if (!(Test-Path $exePath)) {
    Write-Host "error: metrics_agent.exe not found in $PSScriptRoot!" -ForegroundColor Red
    Write-Host "move to folder .exe file in the same folder and try again." -ForegroundColor Yellow
    exit
}
$taskName = "SystemMetricsAgent"

$action = New-ScheduledTaskAction -Execute $exePath
$trigger = New-ScheduledTaskTrigger -AtStartup

Register-ScheduledTask -Action $action -Trigger $trigger -TaskName $taskName -Description "metrix agent" -User "NT AUTHORITY\SYSTEM" -RunLevel Highest

Write-Host "agent succesfull upload" -ForegroundColor Green
Write-Host "exe path: $exePath" -ForegroundColor Cyan
Write-Host "start agent: Start-ScheduledTask -TaskName '$taskName'" -ForegroundColor Yellow