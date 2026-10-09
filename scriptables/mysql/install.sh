#!/bin/bash
# exit-on-failure=yes
set -e

if dpkg -s mariadb-server >/dev/null 2>&1; then
    echo "MariaDB is already installed on this machine. Remove it before installing MySQL."
    exit 1
fi

ROOT_PASSWORD='#ROOT_PASSWORD#'

sudo apt-get update -y
sudo apt-get install -y mysql-server mysql-client

sudo systemctl enable --now mysql

sudo mysql <<SQL
ALTER USER 'root'@'localhost' IDENTIFIED WITH caching_sha2_password BY '$ROOT_PASSWORD';
FLUSH PRIVILEGES;
SQL

sudo install -m 600 /dev/stdin /root/.my.cnf <<CNF
[client]
user=root
password="$ROOT_PASSWORD"
CNF

echo "MySQL root password set. Root can still run mysql without a password through /root/.my.cnf."
