#!/bin/bash
# exit-on-failure=yes

PHP_PATH=$(which php#PHP_VERSION#)
LOG_PATH=/var/log/workers/#SITE_SLUG#

sudo mkdir -p $LOG_PATH
sudo CHOWN -R #SITE_SLUG#:www-data $LOG_PATH



sudo touch /etc/systemd/system/#SITE_SLUG#.service

sudo echo <<-EOF
[Unit]
Description=Queue worker for #SITE_SLUG#

[Service]
User=#SITE_SLUG#
Group=www-data
Restart=always
WorkingDirectory=#USER_DIRECTORY#/#SITE_SLUG#
ExecStart= $PHP_PATH #USER_DIRECTORY#/#SITE_SLUG#/artisan #COMMAND#
StandardOutput=$LOG_PATH/info.log
StandardError=$LOG_PATH/error.log

[Install]
WantedBy=multi-user.target
EOF > /etc/systemd/system/#SITE_SLUG#.service

sudo systemctl daemon-reload
sudo systemctl enable SITE_SLUG#.service
sudo systemctl start SITE_SLUG#.service