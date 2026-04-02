DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

OLD_SERVICES=("server_master" "server_agent" "system_metrics" "master_monitoring")

echo -e "\e[33m--- Installing Cinelink Monitoring (Linux) ---\e[0m\n"

echo -e "\e[90m Cleaning up old services and processes...\e[0m"
for svc in "${OLD_SERVICES[@]}"; do
    if systemctl list-unit-files | grep -q "$svc.service"; then
        sudo systemctl stop "$svc.service" >/dev/null 2>&1
        sudo systemctl disable "$svc.service" >/dev/null 2>&1
        sudo rm "/etc/systemd/system/$svc.service" >/dev/null 2>&1
        echo -e "  - Old service '$svc' removed."
    fi
done

sudo pkill -f "master_server" >/dev/null 2>&1
sudo pkill -f "metrics_agent" >/dev/null 2>&1
sudo systemctl daemon-reload

register_service() {
    local name=$1
    local binary=$2
    local description=$3
    local path="$DIR/$binary"

    if [ ! -f "$path" ]; then
        echo -e "\e[31m Error: File $binary not found in $DIR. Skipping...\e[0m"
        return
    fi

    echo -e "\e[33m Setting up service for $name...\e[0m"
    chmod +x "$path"

    sudo bash -c "cat > /etc/systemd/system/$name.service" << EOL
[Unit]
Description=$description
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$DIR
ExecStart=$path
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOL

    sudo systemctl daemon-reload
    sudo systemctl enable "$name.service"
    sudo systemctl restart "$name.service"
    echo -e "\e[32m Сервіс $name успішно запущено!\e[0m"
}

register_service "server_master" "master_server" "Master Monitoring Server"
register_service "server_agent" "metrics_agent" "System Metrics Agent"

echo -e "\n\e[33m Checking local port 8082...\e[0m"
sleep 3 

if sudo ss -tulpn | grep -q ":8082"; then
    echo -e "\e[36m Port 8082 is active! Opening browser with test...\e[0m"
    
    URL="http://127.0.0.1:8082/system-metrics"
    
    if command -v xdg-open > /dev/null; then
        xdg-open "$URL"
    elif command -v open > /dev/null; then
        open "$URL"
    else
        echo -e "\e[35m Browser not found. Please navigate to the URL manually: $URL\e[0m"
    fi
else
    echo -e "\e[31m Error: Port 8082 is not responding. Check status: sudo systemctl status server_master\e[0m"
fi

echo -e "\n\e[32m All done!\e[0m"