sudo apt-get install mysql-server mysql-client -y
sudo systemctl start mysql

unset HISTFILE

sudo mysql -e "DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');"
sudo mysql -e "DELETE FROM mysql.user WHERE User='';"
sudo mysql -e "DROP DATABASE test;DELETE FROM mysql.db WHERE Db='test' OR Db='test\\_%';"
sudo mysql -e "UPDATE mysql.user SET Password=PASSWORD('#MYSQL_ROOT_PASSWORD#') WHERE User='root';"
sudo mysql -e "ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY '#MYSQL_ROOT_PASSWORD#';"
sudo mysql -e "FLUSH PRIVILEGES;"

sudo systemctl restart mysql