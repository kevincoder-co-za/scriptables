#!/bin/bash
# exit-on-failure=yes

unset HISTFILE

if ! mysql -uroot --password="#MYSQL_ROOT_PASSWORD#" -e scriptable -e "select * from servers limit 1"; then
    mysql -uroot --password="#MYSQL_ROOT_PASSWORD#" -e "CREATE USER #SITE_SLUG#@localhost identified by '#MYSQL_PASSWORD#';"
    mysql -uroot --password="#MYSQL_ROOT_PASSWORD#" -e "CREATE DATABASE #SITE_SLUG#;"
    mysql -uroot --password="#MYSQL_ROOT_PASSWORD#" -e "GRANT ALL PRIVILEGES ON #SITE_SLUG#.* TO #SITE_SLUG#@localhost;" 
    mysql -uroot --password="#MYSQL_ROOT_PASSWORD#" -e "GRANT SELECT PRIVILEGES ON information_schema.* TO #SITE_SLUG#@localhost;" 
    mysql -uroot --password="#MYSQL_ROOT_PASSWORD#" -e "FLUSH PRIVILEGES;"
fi
