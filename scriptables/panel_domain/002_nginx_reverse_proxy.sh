#!/bin/bash
# exit-on-failure=yes
set -e

sudo tee /etc/nginx/sites-available/scriptables >/dev/null <<CONF
server {
    listen 80;
    server_name #DOMAIN#;

    location / {
        proxy_pass http://127.0.0.1:#PANEL_PORT#;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 300s;
    }
}
CONF

sudo ln -sf /etc/nginx/sites-available/scriptables /etc/nginx/sites-enabled/scriptables
sudo nginx -t
sudo systemctl reload nginx
