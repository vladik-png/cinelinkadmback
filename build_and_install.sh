echo -e "\e[33m Starting compilation for Linux...\e[0m"

go mod tidy

go build -o master_server main.go
if [ $? -ne 0 ]; then
    echo -e "\e[31m Error building main.go\e[0m"
    exit 1
fi

go build -o metrics_agent ./agent/metrix.go
if [ $? -ne 0 ]; then
    echo -e "\e[31m Error compiling agent/metrix.go\e[0m"
    exit 1
fi

echo -e "\e[32m Compilation completed successfully!\e[0m"

chmod +x install_full.sh
sudo ./install_full.sh