#!/bin/bash
# exit-on-failure=yes

sudo apt install -y redis-server
sudo systemctl stop redis.service

sudo sed -i 's/^bind .*/bind 127.0.0.1/' /etc/redis/redis.conf

sudo systemctl start redis.service
