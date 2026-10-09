#!/bin/bash
# exit-on-failure=yes
set -e

NOTIFY_EMAIL='#NOTIFY_EMAIL#'

EMAIL_OPTION="--register-unsafely-without-email"
if [ -n "$NOTIFY_EMAIL" ]; then
    EMAIL_OPTION="--email $NOTIFY_EMAIL"
fi

sudo certbot --nginx --non-interactive --agree-tos --redirect $EMAIL_OPTION -d '#DOMAIN#'
