#!/bin/bash
# exit-on-failure=yes

cd #USER_DIRECTORY#/#SITE_SLUG#

echo "Installing composer packages..."

sudo php#PHP_VERSION# /usr/bin/composer.phar install --no-dev
sudo php#PHP_VERSION# /usr/bin/composer.phar dump-autoload

echo "Clearing caches..."

sudo php#PHP_VERSION# artisan config:clear
sudo php#PHP_VERSION# artisan cache:clear
sudo php#PHP_VERSION# artisan view:clear

echo "Fix directory permissions..."
sudo chown -R #SITE_SLUG#:www-data /home/#SITE_SLUG#/#SITE_SLUG#

echo "Linking storage folder..."

sudo -u #SITE_SLUG# php#PHP_VERSION# artisan storage:link --quiet

echo "Setup done."