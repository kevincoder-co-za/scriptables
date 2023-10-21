#!/bin/bash
# exit-on-failure=yes

sudo ufw allow https
sudo ufw allow http
SCRIPTABLE::IMPORT security_poststeps

rm -rf /etc/nginx/sites-enabled/default