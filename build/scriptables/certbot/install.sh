#!/bin/bash
# exit-on-failure=yes

if ! command -v snap &> /dev/null; then
    sudo apt-get update -y
    sudo apt-get install snapd -y
fi

sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot

sudo bash -c 'echo "30 4,16 * * * root /usr/bin/certbot renew --quiet" > /etc/cron.d/certbot-renew'
sudo service cron restart