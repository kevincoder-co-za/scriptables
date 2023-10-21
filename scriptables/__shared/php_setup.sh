sudo add-apt-repository -y ppa:ondrej/php

sudo apt-get update -y

sudo apt-get install -y php#PHP_VERSION#-cli php#PHP_VERSION#-fpm php#PHP_VERSION#-mysql \
php#PHP_VERSION#-curl php#PHP_VERSION#-gd php#PHP_VERSION#-mbstring php#PHP_VERSION#-xml \
php#PHP_VERSION#-zip php#PHP_VERSION#-bcmath php#PHP_VERSION#-memcache php#PHP_VERSION#-redis \
php#PHP_VERSION#-imagick php#PHP_VERSION#-intl php#PHP_VERSION#-ldap php#PHP_VERSION#-soap php#PHP_VERSION#-xmlrpc \
php#PHP_VERSION#-mongodb php#PHP_VERSION#-pgsql php#PHP_VERSION#-sqlite3

sudo sed -i "s/^listen = .*/listen = 127.0.0.1:#FPM_PORT#/" /etc/php/#PHP_VERSION#/fpm/pool.d/www.conf

sudo service php#PHP_VERSION#-fpm stop || true
sudo service php#PHP_VERSION#-fpm start

sudo mkdir -p /var/www/
sudo chown -R www-data:www-data /var/www/

if command -v "composer" > /dev/null 2>&1; then
    exit 0
fi

cd /tmp/ && wget https://getcomposer.org/download/latest-stable/composer.phar
sudo mv /tmp/composer.phar /usr/bin/composer.phar
sudo chmod +x /usr/bin/composer.phar