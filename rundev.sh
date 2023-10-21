#!/bin/bash
set -e

# This script will launch a dev version of Scriptables.
# Please setup a local mysql DB and restore the contents from build/docker-entrypoint-initdb.d
# You will also need to run redis locally.

export MYSQL_HOST=127.0.0.1
export MYSQL_PORT=3306
export MYSQL_DATABASE="scriptable"
export MYSQL_USER="kevin"
export MYSQL_PASSWORD=1234
export MYSQL_ROOT_PASSWORD=1234
export SCRIPTABLE_URL=http://127.0.0.1:3001
export ALLOWED_IPS=127.0.0.1
export REDIS_DSN=127.0.0.1:6379
export ENCRYPTION_KEY="*~#^7^#s0^Z)^^7%b89@#$%5"
export SMTP_HOST=sandbox.smtp.mailtrap.io
export SMTP_USERNAME=xxxx
export SMTP_PASSWORD=xxx
export SMTP_PORT=586
export ALLOW_REGISTER=true
export SMTP_FROM_EMAIL="Scriptables <noreply@test.com>"
export TZ=Africa/Johannesburg
export SCRIPTABLES_SERVER_DSN_HOST=0.0.0.0
export SCRIPTABLES_SERVER_DSN_PORT=3001
export VERBOSE_LOGS=yes
export GIN_MODE=release

air