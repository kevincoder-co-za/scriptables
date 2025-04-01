#!/bin/sh

if [ ! -d "./scriptables" ]; then
  git clone https://github.com/plexcorp-pty-ltd/scriptables.git scriptables
  cd scriptables
fi

if ! command -v docker >/dev/null; then
  curl -fsSL https://get.docker.com -o get-docker.sh
  sh get-docker.sh
  rm get-docker.sh
  usermod -aG docker "$USER"
fi


USER_TIMEZONE=$(cat /etc/timezone)
MYSQL_USER_PASSWORD=$(date +%s | sha256sum | base64 | head -c 15)
MYSQL_ROOT_PASSWORD=$(date +%s | sha256sum | base64 | head -c 15)
ENCRYPTION_KEY=$(date +%s | sha256sum | base64 | head -c 16)

echo "
MYSQL_HOST=scriptables-db
MYSQL_PORT=3306
MYSQL_DATABASE=scriptables
MYSQL_USER=scriptable
MYSQL_PASSWORD=$MYSQL_USER_PASSWORD
MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD
SCRIPTABLE_URL=http://127.0.0.1:3012
ALLOWED_IPS=127.0.0.1
REDIS_DSN=scriptables-redis:6379
ENCRYPTION_KEY=$ENCRYPTION_KEY
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_USERNAME=xxxx
SMTP_PASSWORD=xxx
SMTP_PORT=586
ALLOW_REGISTER=true
SMTP_FROM_EMAIL=Scriptables <noreply@test.com>
TZ=$USER_TIMEZONE
SCRIPTABLES_SERVER_DSN_HOST=0.0.0.0
SCRIPTABLES_SERVER_DSN_PORT=3012
VERBOSE_LOG=yes
GIN_MODE=release
" > .env

docker compose -f docker-compose.yml up -d --build

echo "Install complete. Please visit: http://127.0.0.1:3012/users/register"