# Industrial Protocol Honeypot Deployment Guide

## Overview
This document provides instructions for deploying and testing the Industrial Protocol Honeypot System on CENTOS.

## Prerequisites
- CENTOS 7 or later
- Go 1.21+ (if compiling from source)
- Root or sudo privileges
- libpcap-devel (for packet capture functionality)

## Deployment Steps

### 1. Install Dependencies
```bash
sudo yum install -y libpcap-devel
```

### 2. Transfer the Binary
Copy the compiled binary `honeypot_server_linux` to your CENTOS server.

### 3. Create Configuration Directory
```bash
mkdir -p config logs data/pcap data
```

### 4. Copy Configuration File
Copy the `config/config.yaml` file to the server:

### 5. Run the Honeypot
```bash
# Run in foreground (for testing)
./honeypot_server_linux

# Run in background (for production)
nohup ./honeypot_server_linux > honeypot.log 2>&1 &
```

## Testing the Honeypot

### 1. Check Web Interface
Open a web browser and navigate to:
```
http://<server-ip>:8080
```

### 2. Test Industrial Protocols
Use appropriate client tools to test the honeypot's response to industrial protocol requests:

#### Modbus TCP Test
```bash
# Using mbpoll (install with: sudo yum install -y mbpoll)
mbpoll -t 0 -c 1 -r 1 -p 502 <server-ip>
```

#### MySQL Test
```bash
mysql -h <server-ip> -P 3306 -u root -p
```

#### Redis Test
```bash
redis-cli -h <server-ip> -p 6379 ping
```

#### Kafka Test
```bash
# Using kafka-console-producer
kafka-console-producer --broker-list <server-ip>:9092 --topic test
```

## Monitoring and Management

### 1. Check Logs
```bash
tail -f logs/honeypot.log
```

### 2. View Captured Packets
Captured packets are stored in the `data/pcap` directory.

### 3. API Access
The honeypot provides RESTful API endpoints:

- Get all sessions: `GET /api/sessions`
- Get all fingerprints: `GET /api/fingerprints`
- Get statistics: `GET /api/stats`

## Enabling Packet Capture

To enable packet capture functionality, ensure that:

1. `libpcap-devel` is installed
2. The configuration file has `packet_capture.enabled: true`
3. The binary was compiled with `CGO_ENABLED=1`

## Troubleshooting

### Port Already in Use
If you see "address already in use" errors, check which process is using the port:
```bash
sudo netstat -tulpn | grep <port>
```

### Permission Denied
Ensure the binary has execute permissions:
```bash
chmod +x honeypot_server_linux
```

### Packet Capture Not Working
Check if the user has permission to capture packets:
```bash
# Add user to wireshark group (if exists)
sudo usermod -a -G wireshark $USER

# Or run with sudo
sudo ./honeypot_server_linux
```

## Performance Considerations

- The honeypot will consume more resources when full packet capture is enabled
- Monitor CPU and memory usage on production systems
- Adjust the `session.timeout` and `session.cleanup_interval` settings based on your environment

## Security Considerations

- Run the honeypot in a isolated environment
- Limit network access to the honeypot
- Regularly check logs for suspicious activity
- Update the honeypot software regularly

## Contact

For support or questions, please refer to the project documentation.