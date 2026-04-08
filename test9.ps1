try {
    $port = 9
    $Udp = New-Object System.Net.Sockets.UdpClient($port)
    $IP = [System.Net.IPEndPoint]::new([System.Net.IPAddress]::Any, $port)

    Write-Host "[LISTENER STARTED] Waiting for magic packets on UDP port $port..." -ForegroundColor Cyan
    Write-Host "Press Ctrl+C to stop." -ForegroundColor Gray
    Write-Host "---------------------------------------------------"

    while($true) {
        $Data = $Udp.Receive([ref]$IP)
        $Time = (Get-Date).ToString("HH:mm:ss")
        
        Write-Host "[$Time] BOOM! PACKET RECEIVED!" -ForegroundColor Yellow
        Write-Host "    Sender: $($IP.Address)" -ForegroundColor Green
        Write-Host "    Size: $($Data.Length) bytes" -ForegroundColor Green
        Write-Host "---------------------------------------------------"
    }
}
catch {
    Write-Host "`nERROR: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Port $port might be in use or blocked by firewall." -ForegroundColor Yellow
}
finally {
    if ($null -ne $Udp) { $Udp.Close() }
    Write-Host "`nPress Enter to close..." -ForegroundColor White
    Read-Host
}