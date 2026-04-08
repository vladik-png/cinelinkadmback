#!/bin/bash
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
OLD_SERVICES=("server_master" "server_agent" "system_metrics" "master_monitoring")

draw_progress() {
    local duration=$1
    local sleep_interval=$(echo "scale=2; $duration / 50" | bc)
    echo -ne "  Progress: ["
    for i in {1..50}; do
        echo -ne "#"
        sleep $sleep_interval
    done
    echo -e "] Done!"
}

clear
echo -e "\e[1;36m┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\e[0m"
echo -e "\e[1;36m┃            CINELINK LINUX DEPLOYMENT ENGINE               ┃\e[0m"
echo -e "\e[1;36m┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\e[0m"

echo -e "\n\e[1;33m[1/4] Purging legacy services...\e[0m"
draw_progress 1
for svc in "${OLD_SERVICES[@]}"; do
    if systemctl list-unit-files | grep -q "$svc.service"; then
        sudo systemctl stop "$svc.service" >/dev/null 2>&1
        sudo systemctl disable "$svc.service" >/dev/null 2>&1
        sudo rm "/etc/systemd/system/$svc.service" >/dev/null 2>&1
    fi
done
sudo pkill -f "master_server" >/dev/null 2>&1
sudo pkill -f "metrics_agent" >/dev/null 2>&1
sudo systemctl daemon-reload
echo -e "  \e[32m✔ Cleanup complete.\e[0m"

echo -e "\n\e[1;33m[2/4] Installing new modules...\e[0m"
draw_progress 1.5
register_service() {
    local name=$1
    local binary=$2
    local path="$DIR/$binary"
    if [ -f "$path" ]; then
        chmod +x "$path"
        sudo bash -c "cat > /etc/systemd/system/$name.service" << EOL
[Unit]
Description=$3
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
        sudo systemctl enable "$name.service" >/dev/null 2>&1
        sudo systemctl restart "$name.service"
        echo -e "  \e[32m✔ Service '$name' -> ACTIVE\e[0m"
    else
        echo -e "  \e[1;31m✖ ERROR: Binary '$binary' not found in $DIR\e[0m"
    fi
}

register_service "server_master" "master_server" "Master Monitoring Server"
register_service "server_agent" "metrics_agent" "System Metrics Agent"


echo -e "\n\e[1;33m[3/4] Configuring Linux Firewall (UFW)...\e[0m"
draw_progress 1
if command -v ufw >/dev/null 2>&1; then
    sudo ufw allow 8080/tcp >/dev/null 2>&1
    sudo ufw allow 8081/tcp >/dev/null 2>&1
    sudo ufw allow 9/udp >/dev/null 2>&1
    echo -e "  \e[32m✔ Ports 8080, 8081 (TCP) and 9 (UDP) opened successfully.\e[0m"
else
    echo -e "  \e[33m⚠ UFW is not installed. Skipping firewall configuration.\e[0m"
fi


echo -e "\n\e[1;33m[4/4] Verifying network stack...\e[0m"
draw_progress 2
if sudo ss -tulpn | grep -q ":8080"; then
    echo -e "  \e[1;32m✔ SUCCESS: Port 8080 is ACTIVE!\e[0m"
    # xdg-open "http://127.0.0.1:8080/system-metrics" >/dev/null 2>&1 &
else
    echo -e "  \e[1;31m✖ ERROR: Port 8080 not responding.\e[0m"
fi

echo -e "\n\e[1;36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\e[0m"
echo -e "         Deployment Finished. Systems are Nominal.           "
echo -e "\e[1;36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\e[0m"