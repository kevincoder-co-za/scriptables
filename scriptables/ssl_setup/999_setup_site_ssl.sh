#!/bin/bash
# exit-on-failure=yes

sudo certbot --nginx --non-interactive --agree-tos --email="#NOTIFY_EMAIL#" -d #DOMAIN#