#!/bin/bash
# exit-on-failure=yes

cd #USER_DIRECTORY#/#SITE_SLUG#

echo "Setup .env variables"

MYSQL_PASSWORD=$(cat <<'PASSWORD'
#MYSQL_PASSWORD#
PASSWORD
)
SED_SAFE_MYSQL_PASSWORD=$(printf '%s' "$MYSQL_PASSWORD" | sed -e 's/[\\&|]/\\&/g')

sudo cp #ENVIRONMENT#.env .env
sudo sed -i "s/PLEX_MYSQL_DB/#SITE_SLUG#/g" .env
sudo sed -i "s/PLEX_MYSQL_USERNAME/#SITE_SLUG#/g" .env
sudo sed -i "s|PLEX_MYSQL_PASSWORD|$SED_SAFE_MYSQL_PASSWORD|g" .env
