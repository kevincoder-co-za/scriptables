#!/bin/bash
# exit-on-failure=yes
set -e
unset HISTFILE

if ! id "#SUDO_USERNAME#" >/dev/null 2>&1; then
    echo "Creating user #SUDO_USERNAME#"
    sudo useradd --create-home --shell /bin/bash "#SUDO_USERNAME#"
fi

sudo usermod -aG sudo "#SUDO_USERNAME#"

PASSWORD_HASH='#PASSWORD_HASH#'
if [ -n "$PASSWORD_HASH" ]; then
    echo "Setting the sudo password"
    echo "#SUDO_USERNAME#:$PASSWORD_HASH" | sudo chpasswd -e
fi

PUBLIC_KEY=$(cat <<'KEY'
#PUBLIC_KEY#
KEY
)

HOME_DIRECTORY=$(getent passwd "#SUDO_USERNAME#" | cut -d: -f6)
AUTHORIZED_KEYS="$HOME_DIRECTORY/.ssh/authorized_keys"

sudo install -d -m 700 -o "#SUDO_USERNAME#" -g "#SUDO_USERNAME#" "$HOME_DIRECTORY/.ssh"
sudo touch "$AUTHORIZED_KEYS"

if ! sudo grep -qxF "$PUBLIC_KEY" "$AUTHORIZED_KEYS"; then
    echo "Adding the SSH public key"
    echo "$PUBLIC_KEY" | sudo tee -a "$AUTHORIZED_KEYS" >/dev/null
fi

sudo chown "#SUDO_USERNAME#:#SUDO_USERNAME#" "$AUTHORIZED_KEYS"
sudo chmod 600 "$AUTHORIZED_KEYS"

echo "#SUDO_USERNAME# is ready and needs a password for sudo."
