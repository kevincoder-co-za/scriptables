#!/bin/bash
# exit-on-failure=yes
set -e

sudo apt-get update -y
sudo apt-get install -y nginx

sudo systemctl enable --now nginx
