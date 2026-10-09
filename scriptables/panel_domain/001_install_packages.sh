#!/bin/bash
# exit-on-failure=yes
set -e

sudo apt-get update -y
sudo apt-get install -y nginx certbot python3-certbot-nginx

sudo systemctl enable --now nginx
sudo systemctl enable --now certbot.timer
