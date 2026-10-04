#!/bin/bash
# exit-on-failure=yes
set -e

if ! command -v ufw >/dev/null 2>&1; then
    sudo apt-get update -y
    sudo apt-get install -y ufw
fi

sudo ufw allow #SSH_PORT#/tcp

if ! sudo ufw status | grep -q "Status: active"; then
    for port in $(sudo sshd -T 2>/dev/null | awk '$1 == "port" {print $2}'); do
        sudo ufw allow "$port/tcp"
    done

    sudo ufw allow 80/tcp
    sudo ufw allow 443/tcp

    PANEL_PORT="#PANEL_PORT#"
    if [ -n "$PANEL_PORT" ]; then
        sudo ufw allow "$PANEL_PORT/tcp"
    fi

    sudo ufw default deny incoming
    sudo ufw default allow outgoing
    sudo ufw --force enable
fi

sudo ufw status numbered
