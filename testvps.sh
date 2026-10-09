#!/bin/bash

set -euo pipefail

CONTAINER="scriptables-vps"
IMAGE="jrei/systemd-ubuntu:24.04"
GO_VERSION="1.27.1"
APP_USER="scriptables"
APP_DIR="/opt/scriptables"
APP_PORT="1323"
HTTP_PORT="8081"
ENV_FILE="$APP_DIR/data/testvps.env"
REGISTRATION_TOKEN_FILE="/etc/secrets/scriptables/registration_token"
REGISTRATION_TOKEN=""
REPO_DIR="$(cd "$(dirname "$0")" && pwd)"

log() { echo "==> $*"; }

run_as_root() { docker exec "$CONTAINER" bash -c "$1"; }
run_as_app_user() { docker exec -u "$APP_USER" -e HOME="/home/$APP_USER" "$CONTAINER" bash -c "$1"; }

container_exists() { docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER"; }

start_container() {
    if container_exists; then
        log "Starting the existing $CONTAINER container"
        docker start "$CONTAINER" >/dev/null
        return
    fi

    log "Creating the $CONTAINER container"
    docker run -d --name "$CONTAINER" --hostname "$CONTAINER" \
        --privileged --cgroupns=host \
        -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
        -v "$REPO_DIR:$APP_DIR" \
        -p "127.0.0.1:$APP_PORT:$APP_PORT" \
        -p "127.0.0.1:$HTTP_PORT:80" \
        "$IMAGE" >/dev/null
}

wait_for_systemd() {
    until run_as_root "systemctl is-system-running 2>/dev/null | grep -Eq 'running|degraded'"; do
        sleep 1
    done
}

install_system_packages() {
    log "Installing system packages"
    run_as_root "rm -f /usr/sbin/policy-rc.d"
    run_as_root "export DEBIAN_FRONTEND=noninteractive && apt-get update -y >/dev/null && apt-get install -y ca-certificates curl git sudo openssl cron ufw nginx >/dev/null"
    run_as_root "systemctl enable --now nginx cron >/dev/null 2>&1"
}

create_scriptables_user() {
    local host_uid
    host_uid="$(id -u)"

    run_as_root "
        if id $APP_USER >/dev/null 2>&1; then exit 0; fi
        uid_owner=\$(getent passwd $host_uid | cut -d: -f1)
        if [ -n \"\$uid_owner\" ]; then userdel -r \"\$uid_owner\" >/dev/null 2>&1; fi
        useradd --uid $host_uid --create-home --shell /bin/bash $APP_USER
        printf '%s\n' '$APP_USER ALL=(ALL) NOPASSWD:ALL' 'Defaults:$APP_USER env_keep += \"DEBIAN_FRONTEND\"' > /etc/sudoers.d/$APP_USER
        chmod 440 /etc/sudoers.d/$APP_USER
    "
}

install_go() {
    run_as_root "
        if /usr/local/go/bin/go version 2>/dev/null | grep -q go$GO_VERSION; then exit 0; fi
        arch=\$(dpkg --print-architecture)
        curl -fsSL https://go.dev/dl/go$GO_VERSION.linux-\$arch.tar.gz | tar -C /usr/local -xz
    "
}

install_air() {
    log "Installing Go $GO_VERSION and air"
    install_go
    run_as_app_user "test -x ~/go/bin/air || /usr/local/go/bin/go install github.com/air-verse/air@latest"
}

write_env_file() {
    run_as_app_user "
        mkdir -p $APP_DIR/data
        if [ -f $ENV_FILE ]; then exit 0; fi
        cat > $ENV_FILE <<EOF
SQLITE_PATH=$APP_DIR/data/testvps.db
SCRIPTABLES_SERVER_DSN_HOST=0.0.0.0
SCRIPTABLES_SERVER_DSN_PORT=$APP_PORT
SCRIPTABLE_URL=http://127.0.0.1:$APP_PORT
ALLOWED_IPS=127.0.0.1
ENCRYPTION_KEY=\$(openssl rand -hex 16)
SESSION_SECRET=\$(openssl rand -hex 32)
SMTP_USERNAME=xxxx
TZ=UTC
VERBOSE_LOG=yes
GIN_MODE=debug
EOF
    "
}

create_registration_token() {
    local token
    token="$(openssl rand -hex 16)"

    run_as_root "
        install -d -m 700 -o $APP_USER -g $APP_USER \$(dirname $REGISTRATION_TOKEN_FILE)
        printf '%s' '$token' | sha256sum | awk '{print \$1}' > $REGISTRATION_TOKEN_FILE
        chown $APP_USER:$APP_USER $REGISTRATION_TOKEN_FILE
        chmod 600 $REGISTRATION_TOKEN_FILE
    "

    REGISTRATION_TOKEN="$token"
}

has_registration_token() { run_as_root "test -f $REGISTRATION_TOKEN_FILE"; }
has_database() { run_as_root "test -f $APP_DIR/data/testvps.db"; }

create_first_registration_token() {
    if has_database || has_registration_token; then
        return
    fi

    create_registration_token
}

print_registration_token() {
    if [ -z "$REGISTRATION_TOKEN" ]; then
        return
    fi

    log "Register at http://127.0.0.1:$APP_PORT/users/register"
    log "Registration token: $REGISTRATION_TOKEN"
    log "The token is shown only once and stops working after the first account is created."
}

install_air_service() {
    log "Running the app with air as the scriptables-dev service"
    run_as_root "cat > /etc/systemd/system/scriptables-dev.service <<EOF
[Unit]
Description=Scriptables (air live reload)
After=network.target

[Service]
User=$APP_USER
Group=$APP_USER
Restart=always
RestartSec=3
WorkingDirectory=$APP_DIR
Environment=HOME=/home/$APP_USER
Environment=PATH=/usr/local/go/bin:/home/$APP_USER/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
Environment=SCRIPTABLES_ENV_FILE=$ENV_FILE
Environment=CGO_ENABLED=0
Environment=GOFLAGS=-buildvcs=false
ExecStart=/home/$APP_USER/go/bin/air

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload && systemctl enable scriptables-dev >/dev/null 2>&1 && systemctl restart scriptables-dev"
}

enable_firewall() {
    run_as_root "for port in 22 80 443 $APP_PORT; do ufw allow \$port/tcp >/dev/null; done; ufw --force enable >/dev/null"
}

wait_for_app() {
    log "Waiting for the first build"
    until curl -s -o /dev/null "http://127.0.0.1:$APP_PORT/users/login"; do
        sleep 2
    done
}

print_summary() {
    log "Scriptables is running at http://127.0.0.1:$APP_PORT"
    log "Nginx inside the VPS is reachable at http://127.0.0.1:$HTTP_PORT"
    log "Logs:    ./testvps.sh logs"
    log "Shell:   ./testvps.sh shell"
    log "Destroy: ./testvps.sh destroy"
    print_registration_token
}

create_vps() {
    start_container
    wait_for_systemd
    install_system_packages
    create_scriptables_user
    install_air
    write_env_file
    create_first_registration_token
    install_air_service
    enable_firewall
    wait_for_app
    print_summary
}

case "${1:-up}" in
    up) create_vps ;;
    reset-token) create_registration_token; print_registration_token ;;
    logs) docker exec "$CONTAINER" journalctl -u scriptables-dev -f -n 100 ;;
    shell) docker exec -it "$CONTAINER" bash ;;
    stop) docker stop "$CONTAINER" >/dev/null ;;
    destroy) docker rm -f "$CONTAINER" >/dev/null ;;
    *) echo "Usage: ./testvps.sh [up|reset-token|logs|shell|stop|destroy]"; exit 1 ;;
esac
