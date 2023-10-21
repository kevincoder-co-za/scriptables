#!/bin/bash
# exit-on-failure=yes

cd #USER_DIRECTORY#/#SITE_SLUG#

sudo -u #SITE_SLUG# GIT_SSH_COMMAND="ssh -o StrictHostKeyChecking=no -i #KEY_PATH#" git pull

sudo -u www-data php#PHP_VERSION# artisan config:clear
sudo -u www-data php#PHP_VERSION# artisan cache:clear
sudo -u www-data php#PHP_VERSION# artisan view:clear

sudo -u #SITE_SLUG# php#PHP_VERSION# /usr/bin/composer.phar install

sudo -u #SITE_SLUG# php#PHP_VERSION# artisan migrate