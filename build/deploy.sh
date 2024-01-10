#!/bin/sh -ex

# Deploy Scriptables to your own server, replace variables below with your server details.
cd ../

SSH_KEY_PATH="/home/kevin/.ssh/serverkey"
SSH_PORT=22
SSH_USERNAME=root
SERVER_IP=192.168.1.1

tar  -czf /tmp/build.tar.gz .
scp -i $SSH_KEY_PATH -P $SSH_PORT /tmp/build.tar.gz $SSH_USERNAME@$SERVER_IP:/tmp/

ssh -i $SSH_KEY_PATH -p $SSH_PORT $SSH_USERNAME@$SERVER_IP 'sudo systemctl stop scriptables.service || true'
ssh -i $SSH_KEY_PATH -p $SSH_PORT $SSH_USERNAME@$SERVER_IP 'sudo rm -rf /var/www/scriptables/ || true'
ssh -i $SSH_KEY_PATH -p $SSH_PORT $SSH_USERNAME@$SERVER_IP 'sudo mkdir -p /var/www/scriptables/'
ssh -i $SSH_KEY_PATH -p $SSH_PORT $SSH_USERNAME@$SERVER_IP 'cd /var/www/scriptables/ && sudo tar -xzf /tmp/build.tar.gz'
ssh -i $SSH_KEY_PATH -p $SSH_PORT $SSH_USERNAME@$SERVER_IP 'cd /var/www/scriptables/ && sudo mv ./build/systemd/scriptables.service /etc/systemd/system/'
ssh -i $SSH_KEY_PATH -p  $SSH_PORT $SSH_USERNAME@$SERVER_IP 'sudo chown -R www-data:www-data /var/www/scriptables/'
ssh -i $SSH_KEY_PATH -p $SSH_PORT $SSH_USERNAME@$SERVER_IP 'sudo systemctl daemon-reload && sudo systemctl restart scriptables.service'

rm -rf /tmp/build.tar.gz
