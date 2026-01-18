#!/bin/bash

# Industrial Protocol Honeypot Stop Script

# Check if PID file exists
if [ ! -f "honeypot.pid" ]; then
    echo "Error: honeypot.pid file not found!"
    echo "The honeypot may not be running."
    exit 1
fi

PID=$(cat honeypot.pid)
echo "Stopping honeypot with PID: $PID..."

# Try to kill the process gracefully
kill $PID 2>/dev/null

# Wait for process to exit
count=0
while kill -0 $PID 2>/dev/null && [ $count -lt 10 ]; do
    sleep 1
    count=$((count+1))
done

# Force kill if still running
if kill -0 $PID 2>/dev/null; then
    echo "Force killing honeypot process..."
    kill -9 $PID 2>/dev/null
    sleep 1
fi

# Check if process is still running
if kill -0 $PID 2>/dev/null; then
    echo "Error: Failed to stop honeypot process!"
    exit 1
else
    # Remove PID file
    rm -f honeypot.pid
    echo "Honeypot stopped successfully!"
fi
