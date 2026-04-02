#!/bin/bash
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
AGENT_PATH="$DIR/metrics_agent"

if [ ! -f "$AGENT_PATH" ]; then
    echo "error file metrics_agent not found in $DIR!"
    echo "Make sure you compile the program and put it next to the script."
    exit 1
fi

chmod +x "$AGENT_PATH"

SERVICE_FILE="/etc/systemd/system/metrics_agent.service"

echo "Створення сервісу $SERVICE_FILE..."

sudo bash -c "cat > $SERVICE_FILE" << EOL
[Unit]
Description=System Metrics Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$DIR
ExecStart=$AGENT_PATH
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOL

sudo systemctl daemon-reload

sudo systemctl enable metrics_agent.service

sudo systemctl start metrics_agent.service

echo "Agent successfully added to autostart as a systemd service!"
echo "To check the status, run: sudo systemctl status metrics_agent"