#!/bin/bash
# exit-on-failure=yes

sudo apt install memcached -y
sudo apt install libmemcached-tools -y
sudo systemctl start memcached

MEMCACHED_CONF="/etc/memcached.conf"
sudo sed -i '/^-l/s/^/#/g' "$MEMCACHED_CONF"
sudo echo "-l 127.0.0.1" >> "$MEMCACHED_CONF"

sudo systemctl restart memcached