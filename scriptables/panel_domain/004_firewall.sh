#!/bin/bash
set -e

if ! sudo ufw status | grep -q "Status: active"; then
    echo "ufw is not active, leaving the firewall alone."
    exit 0
fi

sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw delete allow '#PANEL_PORT#/tcp' || true

echo "Port #PANEL_PORT# is now closed to the internet. The panel is served through Nginx on https://#DOMAIN#"
