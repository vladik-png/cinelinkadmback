# Cinelink Unified Backend

This is the central monolithic backend and monitoring agent for the Cinelink Admin Panel. It handles metrics collection, server monitoring, alerts, and exposes the REST API on **Port 8081**.

## Running via Docker

You can easily run the entire backend using Docker. This avoids needing to install Go or configure system services manually on your host.

### 1. Build the Docker Image
First, build the image from the root directory of the project:
```bash
docker build -t cinelink-backend .
```

### 2. Run the Container
Start the container and map the required ports. The backend uses port `8081` for the HTTP API and port `9` (UDP) for Wake-on-LAN packets.

```bash
docker run -d \
  --name cinelink-backend \
  -p 8081:8081 \
  -p 9:9/udp \
  cinelink-backend
```

### 3. Check the Logs
To ensure the backend and the internal agent worker started successfully:
```bash
docker logs -f cinelink-backend
```

### Advanced: Host Networking (Recommended for Agents)
If you want the internal agent to correctly capture the **host's** system metrics (CPU, RAM, Disk) instead of the container's isolated resources, run it using the `--network host` flag (Linux only):
```bash
docker run -d \
  --name cinelink-backend \
  --network host \
  cinelink-backend
```