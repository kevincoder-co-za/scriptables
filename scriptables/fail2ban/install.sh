#!/bin/bash
# exit-on-failure=yes
set -e

sudo apt-get update -y
sudo apt-get install -y fail2ban

SSH_PORTS=$(sudo sshd -T 2>/dev/null | awk '$1 == "port" {print $2}' | paste -sd, -)

sudo tee /etc/fail2ban/jail.d/scriptables-sshd.conf >/dev/null <<JAIL
[sshd]
enabled = true
port    = ${SSH_PORTS:-ssh}
backend = systemd
maxretry = 3
bantime  = 1h
findtime = 10m
JAIL

sudo systemctl enable fail2ban
sudo systemctl restart fail2ban
