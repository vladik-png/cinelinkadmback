try {
    $port = 9
    $Udp = New-Object System.Net.Sockets.UdpClient($port)
    $IP = [System.Net.IPEndPoint]::new([System.Net.IPAddress]::Any, $port)

    Clear-Host
    Write-Host "-----------------------------------------------------------" -ForegroundColor Gray
    Write-Host "             CINELINK WoL DIAGNOSTIC LISTENER              " -ForegroundColor Cyan
    Write-Host "-----------------------------------------------------------" -ForegroundColor Gray
    Write-Host "[ONLINE] Waiting for magic packets on UDP port $port..." -ForegroundColor Yellow
    Write-Host "Press Ctrl+C to terminate the listener." -ForegroundColor DarkGray
    Write-Host "-----------------------------------------------------------" -ForegroundColor Gray

    while($true) {
        $Data = $Udp.Receive([ref]$IP)
        $Time = (Get-Date).ToString("HH:mm:ss")
        
        Write-Host "[$Time] ⚡ MAGIC PACKET DETECTED!" -ForegroundColor Yellow
        Write-Host "    Sender IP   : $($IP.Address)" -ForegroundColor Green
        Write-Host "    Source Port : $($IP.Port)" -ForegroundColor DarkGray
        Write-Host "    Packet Size : $($Data.Length) bytes" -ForegroundColor Green
        Write-Host "-----------------------------------------------------------" -ForegroundColor Gray
    }
}
catch {
    Write-Host "`n[!] ERROR: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Port $port might be blocked, in use by another service, or requires Administrator rights." -ForegroundColor Yellow
}
finally {
    if ($null -ne $Udp) { 
        $Udp.Close() 
        Write-Host "`n[!] UDP Socket closed successfully." -ForegroundColor DarkGray
    }
    Write-Host "Press Enter to exit..." -ForegroundColor White
    Read-Host
}