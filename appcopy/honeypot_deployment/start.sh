#!/bin/bash

# Industrial Protocol Honeypot Startup Script

# Check if the binary exists
if [ ! -f "honeypot_server_linux" ]; then
    echo "Error: honeypot_server_linux binary not found!"
    echo "Please make sure you are in the correct directory."
    exit 1
fi

# Create necessary directories
mkdir -p logs data/pcap data

echo "Starting Industrial Protocol Honeypot..."

# Run the honeypot in background
nohup ./honeypot_server_linux > logs/honeypot.log 2>&1 &

# Save the PID
echo $! > honeypot.pid

echo "Honeypot started successfully!"
echo "PID: $(cat honeypot.pid)"
echo "Logs: logs/honeypot.log"
echo "Web interface: http://$(hostname -I | cut -d' ' -f1):8080"
