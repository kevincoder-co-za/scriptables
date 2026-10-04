#!/bin/bash
# exit-on-failure=yes
set -e

if dpkg -s mariadb-server >/dev/null 2>&1; then
    echo "MariaDB is already installed on this machine. Remove it before installing MySQL."
    exit 1
fi

sudo apt-get update -y
sudo apt-get install -y mysql-server mysql-client

sudo systemctl enable --now mysql
