#!/bin/bash
# exit-on-failure=yes
set -e

sudo apt-get update -y
sudo apt-get install -y memcached libmemcached-tools

MEMCACHED_CONF="/etc/memcached.conf"
sudo sed -i '/^-l/s/^/#/g' "$MEMCACHED_CONF"
echo "-l 127.0.0.1" | sudo tee -a "$MEMCACHED_CONF" >/dev/null

sudo systemctl enable memcached
sudo systemctl restart memcached
