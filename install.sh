#!/bin/bash
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
OLD_SERVICES=("server_master" "server_agent" "server_terminal" "system_metrics" "master_monitoring")

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

do_uninstall() {
    echo -e "\n\e[1;31m[] Purging legacy services & files...\e[0m"
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
    sudo pkill -f "terminal_proxy" >/dev/null 2>&1
    sudo systemctl daemon-reload

    rm -f "$DIR/master_server" "$DIR/metrics_agent" "$DIR/terminal_proxy"

    if command -v ufw >/dev/null 2>&1; then
        sudo ufw delete allow 8080/tcp >/dev/null 2>&1
        sudo ufw delete allow 8081/tcp >/dev/null 2>&1
        sudo ufw delete allow 8085/tcp >/dev/null 2>&1
        sudo ufw delete allow 9/udp >/dev/null 2>&1
    fi

    echo -e "  \e[32m✔ Cleanup complete. System is perfectly clean.\e[0m"
    read -p "  Press Enter to continue..."
}

do_install() {
    echo -e "\n\e[1;33m[1/3] Compiling fresh binaries...\e[0m"
    
    echo "  [*] Building Master Server..."
    go build -o master_server .
    if [ $? -ne 0 ]; then echo -e "  \e[1;31m✖ Build failed for Master\e[0m"; read -p "  Press Enter..."; return; fi

    echo "  [*] Building Metrics Agent..."
    go build -o metrics_agent ./cmd/agent
    if [ $? -ne 0 ]; then echo -e "  \e[1;31m✖ Build failed for Agent\e[0m"; read -p "  Press Enter..."; return; fi

    echo -e "\n\e[1;33m[2/3] Registering Systemd Services...\e[0m"
    draw_progress 1.5

    register_service() {
        local name=$1
        local binary=$2
        local desc=$3
        local path="$DIR/$binary"
        if [ -f "$path" ]; then
            chmod +x "$path"
            sudo bash -c "cat > /etc/systemd/system/$name.service" << EOL
[Unit]
Description=$desc
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
        fi
    }

    register_service "server_master" "master_server" "Master Monitoring Server"
    register_service "server_agent" "metrics_agent" "System Metrics Agent"

    echo -e "\n\e[1;33m[3/3] Configuring Linux Firewall (UFW)...\e[0m"
    if command -v ufw >/dev/null 2>&1; then
        sudo ufw allow 8081/tcp >/dev/null 2>&1
        sudo ufw allow 8082/tcp >/dev/null 2>&1
        sudo ufw allow 9/udp >/dev/null 2>&1
        echo -e "  \e[32m✔ Ports 8081, 8082 (TCP) and 9 (UDP) opened successfully.\e[0m"
    else
        echo -e "  \e[33m⚠ UFW is not installed. Skipping firewall configuration.\e[0m"
    fi

    echo -e "\n\e[1;36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\e[0m"
    echo -e "           Deployment Finished. Systems are Nominal.           "
    echo -e "\e[1;36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\e[0m"
    read -p "  Press Enter to return to menu..."
}

while true; do
    clear
    echo -e "\e[1;36m┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\e[0m"
    echo -e "\e[1;36m┃            CINELINK LINUX DEPLOYMENT ENGINE               ┃\e[0m"
    echo -e "\e[1;36m┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\e[0m"
    echo -e "  \e[1;32m[1]\e[0m Install / Update"
    echo -e "  \e[1;33m[2]\e[0m Repair / Reinstall"
    echo -e "  \e[1;31m[3]\e[0m Uninstall"
    echo -e "  \e[1;37m[4]\e[0m Exit"
    echo -e "\e[1;36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\e[0m"
    read -p "  Оберіть дію [1-4]: " choice

    case $choice in
        1) do_install ;;
        2) do_uninstall; do_install ;;
        3) do_uninstall ;;
        4) clear; exit 0 ;;
        *) echo -e "\e[31m  Невірний вибір!\e[0m"; sleep 1 ;;
    esac
done