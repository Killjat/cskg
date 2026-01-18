# Industrial Protocol Honeypot Deployment Package

This package contains all the necessary files to deploy the Industrial Protocol Honeypot on CentOS systems.

## Contents

- `honeypot_server_linux`: Compiled binary for Linux systems
- `config/`: Configuration files
- `web/`: Web interface templates and static files
- `deploy_guide.md`: Detailed deployment instructions
- `start.sh`: Startup script
- `stop.sh`: Stop script

## Quick Start

1. Extract the package to your desired location
2. Run the deployment script:
   ```bash
   chmod +x deploy.sh
   ./deploy.sh
   ```
3. Access the web interface at `http://<server-ip>:8080`

## Manual Deployment

1. Install dependencies:
   ```bash
   sudo yum install -y libpcap-devel
   ```

2. Create necessary directories:
   ```bash
   mkdir -p logs data/pcap data
   ```

3. Run the honeypot:
   ```bash
   ./start.sh
   ```

## Configuration

Edit the `config/config.yaml` file to customize the honeypot settings.

## Monitoring

Check the logs in the `logs/` directory for system activity.

## For more information, please refer to the `deploy_guide.md` file.