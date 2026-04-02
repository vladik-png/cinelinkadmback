DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

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
    
    echo -e "\e[32m Service $name started successfully!\e[0m"
}

register_service "server_master" "master_server" "Master Monitoring Server"
register_service "server_agent" "metrics_agent" "System Metrics Agent"

echo -e "\n\e[36m All available modules installed and running in the background!\e[0m"
echo "Status check: sudo systemctl status server_master"
echo "Status check: sudo systemctl status server_agent"