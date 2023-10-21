#!/bin/bash
# exit-on-failure=yes
# generates-pub-key

unset HISTFILE

echo "Adding user : #username#"

sudo useradd -m -s /bin/bash #username#

if ! sudo getent passwd "#username#" >/dev/null; then
    echo "Error: User #username# does not exist."
    exit 1
fi

echo '#username# ALL=(ALL) NOPASSWD: ALL' | sudo tee -a /etc/sudoers
sudo usermod -aG sudo #username#

echo "Turn off SSH password authentication"
sudo sed -i 's/.*PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo service ssh restart

echo "Setup authorized key for #username#"
mkdir -p /home/#username#/.ssh
touch /home/#username#/.ssh/authorized_keys

echo "#PUBKEY#" >> /home/#username#/.ssh/authorized_keys
eval `ssh-agent` && ssh-keygen -q -N "" -t rsa -f /home/#username#/.ssh/id_rsa

chmod 600 /home/#username#/.ssh/authorized_keys
chmod 700 /home/#username#/.ssh
chmod 600 /home/#username#/.ssh/id_rsa
chmod 644 /home/#username#/.ssh/id_rsa.pub

chown -R #username#:#username# /home/#username#

mkdir -p /var/www/
chown -R www-data:www-data /var/www/