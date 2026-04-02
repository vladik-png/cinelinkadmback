echo -e "\e[33m--- Starting Build & Install (Linux) ---\e[0m"

echo "Building Master..."
go build -o master_server main.go
if [ $? -ne 0 ]; then echo "Build failed for Master"; exit 1; fi

echo "Building Agent..."
go build -o metrics_agent ./agent/metrix.go
if [ $? -ne 0 ]; then echo "Build failed for Agent"; exit 1; fi

chmod +x install_full.sh
sudo ./install_full.sh