echo -e "\e[33m--- Starting Build & Install (Linux) ---\e[0m"

echo "Building Master..."
go build -o master_server .
if [ $? -ne 0 ]; then echo -e "\e[31mBuild failed for Master\e[0m"; exit 1; fi

echo "Building Agent..."
go build -o metrics_agent ./agent/metrix.go
if [ $? -ne 0 ]; then echo -e "\e[31mBuild failed for Agent\e[0m"; exit 1; fi

chmod +x install_full.sh
sudo ./install_full.sh