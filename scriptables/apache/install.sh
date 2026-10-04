#!/bin/bash
# exit-on-failure=yes
set -e

sudo apt-get update -y
sudo apt-get install -y apache2 || true

if systemctl is-active --quiet nginx; then
    sudo sed -i 's/^Listen 80$/Listen 8080/' /etc/apache2/ports.conf
    sudo sed -i 's/<VirtualHost \*:80>/<VirtualHost *:8080>/' /etc/apache2/sites-available/000-default.conf
fi

sudo dpkg --configure -a
sudo systemctl enable apache2
sudo systemctl restart apache2
