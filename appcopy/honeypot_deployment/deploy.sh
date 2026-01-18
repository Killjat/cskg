#!/bin/bash

# Industrial Protocol Honeypot Deployment Script

echo "========================================"
echo "Industrial Protocol Honeypot Deployment"
echo "========================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "This script requires root privileges. Please run with sudo."
    exit 1
fi

# Update system packages
echo "Updating system packages..."
yum update -y > /dev/null

# Install dependencies
echo "Installing dependencies..."
yum install -y libpcap-devel > /dev/null

# Create honeypot user (if not exists)
echo "Creating honeypot user..."
if ! id -u honeypot > /dev/null 2>&1; then
    useradd -r -s /bin/false honeypot
fi

# Create deployment directory
echo "Creating deployment directory..."
DEPLOY_DIR="/opt/honeypot"
mkdir -p $DEPLOY_DIR

# Copy files to deployment directory
echo "Copying files to deployment directory..."
cp -r * $DEPLOY_DIR

# Set permissions
echo "Setting permissions..."
chown -R honeypot:honeypot $DEPLOY_DIR
chmod +x $DEPLOY_DIR/start.sh
chmod +x $DEPLOY_DIR/stop.sh
chmod +x $DEPLOY_DIR/honeypot_server_linux

# Create systemd service file
echo "Creating systemd service..."
cat > /etc/systemd/system/honeypot.service << EOF
[Unit]
Description=Industrial Protocol Honeypot
After=network.target

[Service]
Type=forking
User=honeypot
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/start.sh
ExecStop=$DEPLOY_DIR/stop.sh
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Start and enable service
echo "Starting honeypot service..."
systemctl start honeypot
systemctl enable honeypot

echo "========================================"
echo "Deployment completed successfully!"
echo "========================================"
echo "Service status: $(systemctl status honeypot --no-pager | grep Active)"
echo "Web interface: http://$(hostname -I | cut -d' ' -f1):8080"
echo "Logs: $DEPLOY_DIR/logs/honeypot.log"
echo ""
echo "To manage the service:"
echo "  Start: systemctl start honeypot"
echo "  Stop: systemctl stop honeypot"
echo "  Restart: systemctl restart honeypot"
echo "  Status: systemctl status honeypot"
