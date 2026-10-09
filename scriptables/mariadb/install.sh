#!/bin/bash
# exit-on-failure=yes
set -e

if dpkg -s mysql-server >/dev/null 2>&1; then
    echo "MySQL is already installed on this machine. Remove it before installing MariaDB."
    exit 1
fi

ROOT_PASSWORD='#ROOT_PASSWORD#'

sudo apt-get update -y
sudo apt-get install -y mariadb-server mariadb-client

sudo systemctl enable --now mariadb

sudo mysql <<SQL
ALTER USER 'root'@'localhost' IDENTIFIED VIA unix_socket OR mysql_native_password USING PASSWORD('$ROOT_PASSWORD');
FLUSH PRIVILEGES;
SQL

echo "MariaDB root password set. Root can still run mysql without a password over the unix socket."
