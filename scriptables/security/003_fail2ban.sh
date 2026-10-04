#!/bin/bash
# exit-on-failure=yes
set -e

if ! command -v fail2ban-client >/dev/null 2>&1; then
    sudo apt-get update -y
    sudo apt-get install -y fail2ban
fi

sudo tee /etc/fail2ban/jail.d/scriptables-sshd.conf >/dev/null <<JAIL
[sshd]
enabled = true
port    = #SSH_PORT#
backend = systemd
maxretry = 3
bantime  = 1h
findtime = 10m
JAIL

sudo systemctl enable fail2ban
sudo systemctl restart fail2ban

for attempt in 1 2 3 4 5; do
    if sudo fail2ban-client status sshd; then
        exit 0
    fi
    sleep 2
done

echo "Fail2ban did not start the sshd jail."
exit 1
