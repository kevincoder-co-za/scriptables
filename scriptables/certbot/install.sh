#!/bin/bash
# exit-on-failure=yes
set -e

sudo apt-get update -y
sudo apt-get install -y certbot python3-certbot-nginx

sudo systemctl enable --now certbot.timer
