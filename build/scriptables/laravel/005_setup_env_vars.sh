#!/bin/bash
# exit-on-failure=yes

cd #USER_DIRECTORY#/#SITE_SLUG#

echo "Setup .env variables"

sudo cp #ENVIRONMENT#.env .env
sudo sed -i "s/PLEX_MYSQL_DB/#SITE_SLUG#/g" .env
sudo sed -i "s/PLEX_MYSQL_USERNAME/#SITE_SLUG#/g" .env
sudo sed -i "s/PLEX_MYSQL_PASSWORD/#MYSQL_PASSWORD#/g" .env
