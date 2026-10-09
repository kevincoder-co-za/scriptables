#!/bin/bash

set -euo pipefail

GO_VERSION="1.27.1"
REPO_URL="${SCRIPTABLES_REPO:-https://github.com/kevincoder-co-za/scriptables.git}"
APP_USER="scriptables"
APP_DIR="/opt/scriptables"
APP_PORT="1323"
ENV_FILE="$APP_DIR/.env"
GO_BINARY="/usr/local/go/bin/go"
REGISTRATION_TOKEN_FILE="/etc/secrets/scriptables/registration_token"

DOMAIN=""
CERTIFICATE_EMAIL=""
SKIP_FIREWALL="no"
RESET_REGISTRATION_TOKEN="no"
REGISTRATION_TOKEN=""

log() { echo "==> $*"; }
warn() { echo "Warning: $*" >&2; }
die() { echo "Error: $*" >&2; exit 1; }

show_usage() {
    cat <<EOF
Usage: autoprovision.sh [--domain example.com] [--email you@example.com] [--skip-firewall] [--reset-registration-token]

  --domain          Serve Scriptables over HTTPS on this domain through an Nginx reverse proxy.
  --email           Email address for Let's Encrypt expiry notices. Only used with --domain.
  --skip-firewall   Leave ufw untouched.
  --reset-registration-token
                    Issue a new one time token for registering the first user.
EOF
}

parse_arguments() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --domain) DOMAIN="${2:-}"; shift 2 ;;
            --domain=*) DOMAIN="${1#*=}"; shift ;;
            --email) CERTIFICATE_EMAIL="${2:-}"; shift 2 ;;
            --email=*) CERTIFICATE_EMAIL="${1#*=}"; shift ;;
            --skip-firewall) SKIP_FIREWALL="yes"; shift ;;
            --reset-registration-token) RESET_REGISTRATION_TOKEN="yes"; shift ;;
            -h|--help) show_usage; exit 0 ;;
            *) show_usage; die "Unknown option: $1" ;;
        esac
    done

    if [ -n "$DOMAIN" ] && ! [[ "$DOMAIN" =~ ^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$ ]]; then
        die "\"$DOMAIN\" is not a valid domain name."
    fi
}

require_root() {
    [ "$(id -u)" -eq 0 ] || die "Please run this script as root, e.g. with sudo."
}

require_supported_system() {
    command -v apt-get >/dev/null || die "Only Debian and Ubuntu based systems are supported."
    command -v systemctl >/dev/null || die "systemd is required."
}

install_system_packages() {
    log "Installing system packages"
    export DEBIAN_FRONTEND=noninteractive

    local packages="ca-certificates curl git sudo openssl cron ufw nginx"
    if [ -n "$DOMAIN" ]; then
        packages="$packages certbot python3-certbot-nginx"
    fi

    apt-get update -y
    apt-get install -y $packages
    systemctl enable --now nginx
}

create_scriptables_user() {
    if id "$APP_USER" >/dev/null 2>&1; then
        return
    fi

    log "Creating the $APP_USER user"
    useradd --system --home-dir "$APP_DIR" --no-create-home --shell /bin/bash "$APP_USER"
}

allow_passwordless_sudo() {
    local sudoers_file="/etc/sudoers.d/$APP_USER"

    cat > "$sudoers_file" <<EOF
$APP_USER ALL=(ALL) NOPASSWD:ALL
Defaults:$APP_USER env_keep += "DEBIAN_FRONTEND"
EOF

    chmod 440 "$sudoers_file"
    visudo -cf "$sudoers_file" >/dev/null || { rm -f "$sudoers_file"; die "Generated an invalid sudoers file."; }
}

find_local_checkout() {
    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]:-}")" 2>/dev/null && pwd || true)"

    if [ -n "$script_dir" ] && [ -f "$script_dir/go.mod" ] && [ -f "$script_dir/main.go" ]; then
        echo "$script_dir"
    fi
}

fetch_source() {
    local local_checkout
    local_checkout="$(find_local_checkout)"

    mkdir -p "$APP_DIR"

    if [ -n "$local_checkout" ] && [ "$local_checkout" != "$APP_DIR" ]; then
        log "Copying $local_checkout into $APP_DIR"
        tar -C "$local_checkout" --exclude=.env --exclude='*.db*' --exclude=./bin --exclude=./data -cf - . | tar -C "$APP_DIR" -xf -
    elif [ -d "$APP_DIR/.git" ]; then
        log "Updating the checkout in $APP_DIR"
        chown -R "$APP_USER:$APP_USER" "$APP_DIR"
        sudo -u "$APP_USER" git -C "$APP_DIR" pull --ff-only || warn "Could not fast forward, continuing with the current checkout."
    elif [ ! -f "$APP_DIR/go.mod" ]; then
        log "Cloning $REPO_URL into $APP_DIR"
        git clone "$REPO_URL" "$APP_DIR"
    fi

    mkdir -p "$APP_DIR/data" "$APP_DIR/bin"
    chown -R "$APP_USER:$APP_USER" "$APP_DIR"
}

installed_go_version() {
    if [ -x "$GO_BINARY" ]; then
        "$GO_BINARY" version | awk '{print $3}' | sed 's/^go//'
    fi
}

is_installed_go_recent_enough() {
    local installed
    installed="$(installed_go_version)"

    [ -n "$installed" ] && [ "$(printf '%s\n%s\n' "$GO_VERSION" "$installed" | sort -V | head -n 1)" = "$GO_VERSION" ]
}

detect_go_architecture() {
    case "$(uname -m)" in
        x86_64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        armv6l|armv7l) echo "armv6l" ;;
        *) die "Unsupported CPU architecture: $(uname -m)" ;;
    esac
}

install_go() {
    if is_installed_go_recent_enough; then
        log "Go $(installed_go_version) is already installed"
        return
    fi

    local archive="go$GO_VERSION.linux-$(detect_go_architecture).tar.gz"

    log "Installing Go $GO_VERSION"
    curl -fsSL "https://go.dev/dl/$archive" -o "/tmp/$archive"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "/tmp/$archive"
    rm -f "/tmp/$archive"
}

build_binary() {
    log "Building the Scriptables binary"
    sudo -u "$APP_USER" env \
        HOME="$APP_DIR" \
        GOPATH="$APP_DIR/.go" \
        GOCACHE="$APP_DIR/.cache/go-build" \
        GOFLAGS="-buildvcs=false" \
        CGO_ENABLED=0 \
        bash -c "cd '$APP_DIR' && '$GO_BINARY' build -o bin/scriptables ."
}

generate_secret() {
    openssl rand -hex "$1"
}

detect_timezone() {
    timedatectl show --property=Timezone --value 2>/dev/null || cat /etc/timezone 2>/dev/null || echo "UTC"
}

detect_server_address() {
    hostname -I 2>/dev/null | awk '{print $1}'
}

listen_host() {
    if [ -n "$DOMAIN" ]; then echo "127.0.0.1"; else echo "0.0.0.0"; fi
}

public_url() {
    if [ -n "$DOMAIN" ]; then echo "https://$DOMAIN"; else echo "http://$(detect_server_address):$APP_PORT"; fi
}

set_env_value() {
    local key="$1" value="$2"

    if grep -q "^$key=" "$ENV_FILE"; then
        sed -i "s|^$key=.*|$key=$value|" "$ENV_FILE"
    else
        echo "$key=$value" >> "$ENV_FILE"
    fi
}

is_first_install() {
    [ ! -f "$ENV_FILE" ]
}

create_registration_token() {
    if ! is_first_install && [ "$RESET_REGISTRATION_TOKEN" != "yes" ]; then
        return
    fi

    local secrets_dir
    secrets_dir="$(dirname "$REGISTRATION_TOKEN_FILE")"

    REGISTRATION_TOKEN="$(generate_secret 16)"

    install -d -m 700 -o "$APP_USER" -g "$APP_USER" "$secrets_dir"
    printf '%s' "$REGISTRATION_TOKEN" | sha256sum | awk '{print $1}' > "$REGISTRATION_TOKEN_FILE"
    chown "$APP_USER:$APP_USER" "$REGISTRATION_TOKEN_FILE"
    chmod 600 "$REGISTRATION_TOKEN_FILE"
}

write_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        log "Writing $ENV_FILE with freshly generated secrets"
        cat > "$ENV_FILE" <<EOF
SQLITE_PATH=$APP_DIR/data/scriptables.db

SCRIPTABLES_SERVER_DSN_PORT=$APP_PORT
ALLOWED_IPS=127.0.0.1

ENCRYPTION_KEY=$(generate_secret 16)
SESSION_SECRET=$(generate_secret 32)

SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=586
SMTP_USERNAME=xxxx
SMTP_PASSWORD=xxxx
SMTP_FROM_EMAIL=Scriptables <noreply@test.com>

TZ=$(detect_timezone)
VERBOSE_LOG=no
GIN_MODE=release
EOF
    fi

    set_env_value SCRIPTABLES_SERVER_DSN_HOST "$(listen_host)"
    set_env_value SCRIPTABLE_URL "$(public_url)"

    chown "$APP_USER:$APP_USER" "$ENV_FILE"
    chmod 600 "$ENV_FILE"
}

install_systemd_service() {
    log "Installing the scriptables systemd service"
    cat > /etc/systemd/system/scriptables.service <<EOF
[Unit]
Description=Scriptables
After=network.target

[Service]
User=$APP_USER
Group=$APP_USER
Restart=always
RestartSec=3
Environment=HOME=$APP_DIR
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/bin/scriptables

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable scriptables.service
    systemctl restart scriptables.service
}

configure_nginx_reverse_proxy() {
    [ -n "$DOMAIN" ] || return 0

    log "Configuring Nginx to proxy $DOMAIN to 127.0.0.1:$APP_PORT"
    cat > /etc/nginx/sites-available/scriptables <<EOF
server {
    listen 80;
    server_name $DOMAIN;

    location / {
        proxy_pass http://127.0.0.1:$APP_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 300s;
    }
}
EOF

    ln -sf /etc/nginx/sites-available/scriptables /etc/nginx/sites-enabled/scriptables
    nginx -t
    systemctl reload nginx
}

request_ssl_certificate() {
    [ -n "$DOMAIN" ] || return 0

    local email_option="--register-unsafely-without-email"
    if [ -n "$CERTIFICATE_EMAIL" ]; then
        email_option="--email $CERTIFICATE_EMAIL"
    fi

    log "Requesting a Let's Encrypt certificate for $DOMAIN"
    if ! certbot --nginx --non-interactive --agree-tos --redirect $email_option -d "$DOMAIN"; then
        warn "Certbot could not issue a certificate. Point $DOMAIN at this server and re-run this script."
    fi
}

detect_ssh_ports() {
    local ports
    ports="$(sshd -T 2>/dev/null | awk '$1 == "port" {print $2}')"
    echo "${ports:-22}"
}

configure_firewall() {
    if [ "$SKIP_FIREWALL" = "yes" ]; then
        return
    fi

    log "Configuring the firewall"

    for port in $(detect_ssh_ports); do
        ufw allow "$port/tcp" >/dev/null
    done

    ufw allow 80/tcp >/dev/null
    ufw allow 443/tcp >/dev/null

    if [ -n "$DOMAIN" ]; then
        ufw delete allow "$APP_PORT/tcp" >/dev/null 2>&1 || true
    else
        ufw allow "$APP_PORT/tcp" >/dev/null
    fi

    ufw --force enable >/dev/null
}

print_summary() {
    log "Scriptables is installed in $APP_DIR and running as the $APP_USER user."

    if [ -z "$REGISTRATION_TOKEN" ]; then
        return
    fi

    log "Visit $(public_url)/users/register to create your admin account."
    log "Registration token: $REGISTRATION_TOKEN"
    log "The token is shown only once and stops working after the first account is created."
}

main() {
    parse_arguments "$@"
    require_root
    require_supported_system

    install_system_packages
    create_scriptables_user
    allow_passwordless_sudo
    fetch_source
    install_go
    build_binary
    create_registration_token
    write_env_file
    install_systemd_service
    configure_nginx_reverse_proxy
    request_ssl_certificate
    configure_firewall
    print_summary
}

main "$@"
