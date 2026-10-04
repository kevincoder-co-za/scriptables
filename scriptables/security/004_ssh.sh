#!/bin/bash
# exit-on-failure=yes
set -e

SSHD_SETTINGS="/etc/ssh/sshd_config.d/00-scriptables.conf"
HOME_DIRECTORY=$(getent passwd "#SUDO_USERNAME#" | cut -d: -f6)

if ! command -v sshd >/dev/null 2>&1; then
    echo "OpenSSH server is not installed on this machine."
    exit 1
fi

if ! sudo test -s "$HOME_DIRECTORY/.ssh/authorized_keys"; then
    echo "#SUDO_USERNAME# has no SSH key, refusing to lock down SSH."
    exit 1
fi

PREVIOUS_PORTS=$(sudo sshd -T | awk '$1 == "port" {print $2}')

if ! grep -qE '^\s*Include\s+/etc/ssh/sshd_config\.d/\*\.conf' /etc/ssh/sshd_config; then
    sudo sed -i '1i Include /etc/ssh/sshd_config.d/*.conf' /etc/ssh/sshd_config
fi

sudo mkdir -p /etc/ssh/sshd_config.d
sudo tee "$SSHD_SETTINGS" >/dev/null <<SETTINGS
Port #SSH_PORT#
PermitRootLogin no
PasswordAuthentication no
SETTINGS

if ! sudo sshd -t; then
    sudo rm -f "$SSHD_SETTINGS"
    echo "The new SSH settings are invalid, nothing was changed."
    exit 1
fi

sudo systemctl daemon-reload
if systemctl is-active --quiet ssh.socket; then
    sudo systemctl restart ssh.socket
fi
sudo systemctl restart ssh

if ! sudo sshd -T | grep -qx "port #SSH_PORT#"; then
    echo "SSH did not move to port #SSH_PORT#."
    exit 1
fi

for port in $PREVIOUS_PORTS; do
    if [ "$port" != "#SSH_PORT#" ]; then
        sudo ufw delete allow "$port/tcp" || true
    fi
done

echo "SSH now listens on port #SSH_PORT#. Root and password logins are disabled."
echo "Log in with: ssh -p #SSH_PORT# #SUDO_USERNAME#@your-server"
