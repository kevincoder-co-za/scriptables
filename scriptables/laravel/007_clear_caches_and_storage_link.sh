#!/bin/bash
# exit-on-failure=yes

cd #USER_DIRECTORY#/#SITE_SLUG#

echo "Clearing caches..."

sudo php#PHP_VERSION# artisan config:clear
sudo php#PHP_VERSION# artisan cache:clear
sudo php#PHP_VERSION# artisan view:clear


echo "Linking storage folder..."
sudo -u #SITE_SLUG# php#PHP_VERSION# artisan storage:link --quiet

echo "Fix directory permissions..."
sudo chown -R #SITE_SLUG#:www-data #USER_DIRECTORY#/#SITE_SLUG#

sudo chown -R www-data:www-data #USER_DIRECTORY#/#SITE_SLUG#/#WEBROOT#
sudo chown -R www-data:www-data #USER_DIRECTORY#/#SITE_SLUG#/storage