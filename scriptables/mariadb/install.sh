#!/bin/bash
# exit-on-failure=yes
set -e

if dpkg -s mysql-server >/dev/null 2>&1; then
    echo "MySQL is already installed on this machine. Remove it before installing MariaDB."
    exit 1
fi

sudo apt-get update -y
sudo apt-get install -y mariadb-server mariadb-client

sudo systemctl enable --now mariadb
