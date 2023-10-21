#!/bin/bash
# exit-on-failure=yes
# run_as root

sudo cat <<EOF | sudo tee /etc/nginx/sites-enabled/#SITE_SLUG#
server {
    listen 80;
    server_name #DOMAIN#;
    root #USER_DIRECTORY#/#SITE_SLUG#/#WEBROOT#;
 
    index index.php;
    location ~ /\.(git|env) {
        deny all;
        return 403;
    }

    location / {
        try_files \$uri \$uri/ /index.php?\$query_string;
    }

    location ~ \.php$ {
        try_files \$uri =404;
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass 127.0.0.1:#FPM_PORT#;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;
        fastcgi_read_timeout 600s;
        include fastcgi_params;
    }
}
EOF

sudo php#PHP_VERSION#-fpm stop
sudo php#PHP_VERSION#-fpm start

sudo /etc/init.d/nginx stop
sudo /etc/init.d/nginx start